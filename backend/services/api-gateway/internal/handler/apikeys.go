package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/internal/httpx"
	"github.com/carboncircuit/backend/services/api-gateway/internal/caller"
)

type apiKeyResponse struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Prefix     string  `json:"prefix"`
	CreatedAt  string  `json:"created_at"`
	LastUsedAt *string `json:"last_used_at"`
	RevokedAt  *string `json:"revoked_at"`
}

type createAPIKeyRequest struct {
	Name string `json:"name" binding:"required,max=80"`
}

func toAPIKeyResponse(key *identityv1.APIKey) apiKeyResponse {
	return apiKeyResponse{
		ID:         key.GetId(),
		Name:       key.GetName(),
		Prefix:     key.GetPrefix(),
		CreatedAt:  key.GetCreatedAt(),
		LastUsedAt: emptyToNil(key.GetLastUsedAt()),
		RevokedAt:  emptyToNil(key.GetRevokedAt()),
	}
}

func (h *Handlers) ListAPIKeys(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	listed, err := h.Identity.ListAPIKeys(c.Request.Context())
	if err != nil {
		h.failAPIKey(c, err)
		return
	}

	keys := make([]apiKeyResponse, 0, len(listed.GetKeys()))
	for _, key := range listed.GetKeys() {
		keys = append(keys, toAPIKeyResponse(key))
	}

	httpx.Data(c, http.StatusOK, map[string]any{"keys": keys})
}

func (h *Handlers) CreateAPIKey(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	var body createAPIKeyRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.Fail(c, httpx.CodeValidation)
		return
	}

	key, present := httpx.IdempotencyKeyFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeIdempotencyKeyRequired)
		return
	}

	issued, err := h.Identity.CreateAPIKey(c.Request.Context(), key, body.Name)
	if err != nil {
		h.failAPIKey(c, err)
		return
	}

	httpx.Data(c, http.StatusCreated, map[string]any{
		"key":           toAPIKeyResponse(issued.GetKey()),
		"presented_key": issued.GetPresentedKey(),
	})
}

func (h *Handlers) RevokeAPIKey(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	key, present := httpx.IdempotencyKeyFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeIdempotencyKeyRequired)
		return
	}

	if err := h.Identity.RevokeAPIKey(
		c.Request.Context(), key, c.Param("keyId"),
	); err != nil {
		h.failAPIKey(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handlers) failAPIKey(c *gin.Context, err error) {
	switch status.Code(err) {
	case codes.NotFound:
		httpx.Fail(c, httpx.CodeResourceNotFound)
	case codes.PermissionDenied:
		httpx.Fail(c, httpx.CodeForbidden)
	case codes.InvalidArgument:
		httpx.Fail(c, httpx.CodeValidation)
	case codes.AlreadyExists:
		httpx.Fail(c, httpx.CodeConflict)
	default:
		h.Logger.Error("api key upstream failed", errorAttributes(c, err)...)
		httpx.Fail(c, httpx.CodeDependencyUnavailable)
	}
}
