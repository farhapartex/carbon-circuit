package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"time"

	"github.com/carboncircuit/backend/internal/auth"
	"github.com/carboncircuit/backend/internal/cache"
	"github.com/carboncircuit/backend/internal/httpx"
	"github.com/carboncircuit/backend/internal/ratelimit"
	"github.com/carboncircuit/backend/internal/servicetoken"
	"github.com/carboncircuit/backend/services/api-gateway/internal/caller"
	"github.com/carboncircuit/backend/services/api-gateway/internal/upstream"
)

type Handlers struct {
	Identity       *upstream.Identity
	Billing        *upstream.Billing
	Provenance     *upstream.Provenance
	ProvenanceRead *upstream.ProvenanceRead
	Denylist       *auth.Denylist
	Resolver       *caller.Resolver
	Logger         *slog.Logger
	Revision       string
}

type RouterOptions struct {
	Identity        *upstream.Identity
	Billing         *upstream.Billing
	Provenance      *upstream.Provenance
	ProvenanceRead  *upstream.ProvenanceRead
	Limiter         *ratelimit.Limiter
	Verifier        httpx.TokenVerifier
	Denylist        httpx.RevocationChecker
	SessionDenylist *auth.Denylist
	Cache           *cache.Client
	SessionInterval time.Duration
	TrustedProxies  []string
	Resolver        *caller.Resolver
	Signer          *servicetoken.Signer
	Logger          *slog.Logger
	Environment     string
	Revision        string
}

func errorAttributes(c *gin.Context, err error) []any {
	return []any{
		slog.Any("error", err),
		slog.String("request_id", httpx.CorrelationID(c)),
		slog.String("path", c.FullPath()),
	}
}

func NewRouter(options RouterOptions) *gin.Engine {
	if options.Environment != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	handlers := &Handlers{
		Identity:       options.Identity,
		Resolver:       options.Resolver,
		Billing:        options.Billing,
		Provenance:     options.Provenance,
		ProvenanceRead: options.ProvenanceRead,
		Denylist:       options.SessionDenylist,
		Logger:         options.Logger,
		Revision:       options.Revision,
	}

	router := gin.New()

	if err := router.SetTrustedProxies(options.TrustedProxies); err != nil {
		options.Logger.Error("trusted proxy list rejected, no forwarded header will be honoured",
			slog.Any("error", err),
		)
		_ = router.SetTrustedProxies(nil)
	}

	if len(options.TrustedProxies) == 0 {
		options.Logger.Warn("no trusted proxies configured, client ip comes from the peer address only")
	}

	router.Use(
		httpx.Correlate(),
		httpx.RecoverPanics(options.Logger),
		httpx.LogRequests(options.Logger),
	)

	router.NoRoute(func(c *gin.Context) {
		httpx.Fail(c, httpx.CodeResourceNotFound)
	})

	router.GET("/healthz", func(c *gin.Context) {
		httpx.Data(c, http.StatusOK, gin.H{
			"service":  "api-gateway",
			"revision": options.Revision,
			"status":   "alive",
		})
	})

	router.GET("/readyz", func(c *gin.Context) {
		if _, err := options.Identity.Ping(c.Request.Context()); err != nil {
			httpx.Fail(c, httpx.CodeDependencyUnavailable)
			return
		}
		httpx.Data(c, http.StatusOK, gin.H{
			"service":  "api-gateway",
			"status":   "ready",
			"identity": "reachable",
		})
	})

	public := router.Group("/v1")
	public.Use(
		httpx.EndpointClass("public_read"),
		httpx.ResourceKey(func(c *gin.Context) string { return c.Param("publicRef") }),
		httpx.RateLimit(options.Limiter, options.Logger, nil),
	)

	public.GET("/plans", handlers.ListPlans)
	public.GET("/track/:publicRef", handlers.TrackBatch)
	public.GET("/identity/ping", handlers.IdentityPing)

	authenticated := router.Group("/v1")
	authenticated.Use(
		httpx.Authenticate(options.Verifier, options.Denylist, options.Logger),
		caller.Stamp(options.Resolver, options.Signer, options.Logger),
		httpx.EndpointClass("authenticated_read"),
		httpx.EndpointClassFor(endpointClassOf),
		httpx.RateLimit(options.Limiter, options.Logger, organizationOf),
		httpx.RequireIdempotencyKey(),
		caller.RecordSessions(
			options.Identity, options.Cache, options.SessionInterval, options.Logger,
		),
	)

	authenticated.GET("/me", handlers.Me)
	authenticated.GET("/organizations/current", handlers.GetOrganization)
	authenticated.POST("/organizations", handlers.CreateOrganization)
	authenticated.POST("/subscriptions", handlers.CreateSubscription)
	authenticated.POST("/treasury/nonce", handlers.IssueTreasuryNonce)
	authenticated.POST("/treasury", handlers.DesignateTreasury)
	authenticated.GET("/members", handlers.ListMembers)
	authenticated.POST("/invitations", handlers.InviteMember)
	authenticated.DELETE("/invitations/:invitationId", handlers.RevokeInvitation)
	authenticated.PATCH("/members/:userId", handlers.ChangeMemberRole)
	authenticated.DELETE("/members/:userId", handlers.RevokeMember)
	authenticated.POST("/invitations/accept", handlers.AcceptInvitation)
	authenticated.GET("/sessions", handlers.ListSessions)
	authenticated.DELETE("/sessions/:sessionId", handlers.RevokeSession)
	authenticated.GET("/api-keys", handlers.ListAPIKeys)
	authenticated.DELETE("/api-keys/:keyId", handlers.RevokeAPIKey)
	authenticated.POST("/api-keys", handlers.CreateAPIKey)
	authenticated.GET("/facilities", handlers.ListFacilities)
	authenticated.POST("/facilities", handlers.CreateFacility)
	authenticated.GET("/facilities/:facilityId", handlers.GetFacility)
	authenticated.GET("/batches", handlers.ListBatches)
	authenticated.POST("/batches", handlers.CreateBatch)
	authenticated.GET("/batches/:batchId", handlers.GetBatch)
	authenticated.GET("/batches/:batchId/checkpoints", handlers.ListCheckpoints)
	authenticated.POST("/batches/:batchId/checkpoints", handlers.LogCheckpoint)
	authenticated.GET(
		"/batches/:batchId/components/:componentBatchId",
		handlers.GetComponentBatch,
	)

	return router
}

func (h *Handlers) IdentityPing(c *gin.Context) {
	response, err := h.Identity.Ping(c.Request.Context())
	if err != nil {
		h.Logger.Error("identity ping failed", errorAttributes(c, err)...)
		httpx.Fail(c, httpx.CodeDependencyUnavailable)
		return
	}

	httpx.Data(c, http.StatusOK, gin.H{
		"service":            response.GetService(),
		"revision":           response.GetRevision(),
		"database_reachable": response.GetDatabaseReachable(),
	})
}

func endpointClassOf(c *gin.Context) string {
	if c.Request.Method == http.MethodPost && c.FullPath() == "/v1/api-keys" {
		return "api_key_creation"
	}
	return ""
}

func organizationOf(c *gin.Context) string {
	resolved, present := caller.ContextFrom(c.Request.Context())
	if !present {
		return ""
	}
	return resolved.OrganizationID
}
