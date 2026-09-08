package caller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/internal/apikey"
	"github.com/carboncircuit/backend/internal/cache"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/httpx"
	"github.com/carboncircuit/backend/internal/servicetoken"
)

const contextCacheKeyPrefix = "apikey:context:"

type KeyValidator interface {
	ValidateAPIKey(
		ctx context.Context,
		presented string,
	) (*identityv1.ValidateAPIKeyResponse, error)
}

type ingestRoute struct {
	method string
	path   string
}

var ingestRoutes = map[ingestRoute]bool{
	{http.MethodPost, "/v1/batches"}:                      true,
	{http.MethodPost, "/v1/batches/:batchId/checkpoints"}: true,
	{http.MethodGet, "/v1/batches"}:                       true,
	{http.MethodGet, "/v1/batches/:batchId"}:              true,
	{http.MethodGet, "/v1/batches/:batchId/checkpoints"}:  true,
}

func PresentedKey(header string) (string, bool) {
	fields := strings.Fields(header)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "bearer") {
		return "", false
	}

	if !strings.HasPrefix(fields[1], apikey.Scheme+"_") {
		return "", false
	}

	return fields[1], true
}

func contextCacheKey(presented string) string {
	digest := sha256.Sum256([]byte(presented))
	return contextCacheKeyPrefix + hex.EncodeToString(digest[:])
}

func StampAPIKey(
	validator KeyValidator,
	client *cache.Client,
	signer *servicetoken.Signer,
	ttl time.Duration,
	logger *slog.Logger,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		presented, present := PresentedKey(c.GetHeader("Authorization"))
		if !present {
			c.Next()
			return
		}

		if !ingestRoutes[ingestRoute{c.Request.Method, c.FullPath()}] {
			logger.Info("api key presented on a portal-only route",
				slog.String("path", c.FullPath()),
				slog.String("request_id", httpx.CorrelationID(c)),
			)
			httpx.Fail(c, httpx.CodeForbidden)
			c.Abort()
			return
		}

		resolved, err := resolveKey(
			c.Request.Context(), validator, client, ttl, presented,
		)
		if err != nil {
			logger.Info("api key rejected",
				slog.String("request_id", httpx.CorrelationID(c)),
				slog.Any("error", err),
			)
			httpx.Fail(c, httpx.CodeUnauthenticated)
			c.Abort()
			return
		}

		token, err := signer.Issue(resolved)
		if err != nil {
			logger.Error("could not issue service token for an api key",
				slog.Any("error", err),
			)
			httpx.Fail(c, httpx.CodeInternal)
			c.Abort()
			return
		}

		stamped := withContext(c.Request.Context(), resolved)
		c.Request = c.Request.WithContext(grpcx.WithServiceToken(stamped, token))
		c.Next()
	}
}

func resolveKey(
	ctx context.Context,
	validator KeyValidator,
	client *cache.Client,
	ttl time.Duration,
	presented string,
) (servicetoken.Caller, error) {
	key := contextCacheKey(presented)

	if raw, found, err := client.GetString(ctx, key); err == nil && found {
		var cached servicetoken.Caller
		if json.Unmarshal([]byte(raw), &cached) == nil {
			return cached, nil
		}
	}

	validated, err := validator.ValidateAPIKey(ctx, presented)
	if err != nil {
		return servicetoken.Caller{}, err
	}

	resolved := servicetoken.Caller{
		Subject:            "apikey|" + validated.GetPrefix(),
		UserID:             validated.GetActingUserId(),
		OrganizationID:     validated.GetOrganizationId(),
		OrganizationName:   validated.GetOrganizationName(),
		OrganizationType:   organizationTypeName[validated.GetOrganizationType()],
		OrganizationState:  organizationStateName[validated.GetOrganizationState()],
		VerificationStatus: verificationStatusName[validated.GetVerificationStatus()],
		Role:               "member",
		Credential:         servicetoken.CredentialAPIKey,
	}

	if ttl > 0 {
		if encoded, marshalErr := json.Marshal(resolved); marshalErr == nil {
			_ = client.SetString(ctx, key, string(encoded), ttl)
		}
	}

	return resolved, nil
}
