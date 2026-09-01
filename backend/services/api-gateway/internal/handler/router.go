package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/carboncircuit/backend/internal/httpx"
	"github.com/carboncircuit/backend/internal/ratelimit"
	"github.com/carboncircuit/backend/services/api-gateway/internal/upstream"
)

type Handlers struct {
	Identity *upstream.Identity
	Billing  *upstream.Billing
	Logger   *slog.Logger
	Revision string
}

type RouterOptions struct {
	Identity    *upstream.Identity
	Billing     *upstream.Billing
	Limiter     *ratelimit.Limiter
	Verifier    httpx.TokenVerifier
	Denylist    httpx.RevocationChecker
	Logger      *slog.Logger
	Environment string
	Revision    string
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
		Identity: options.Identity,
		Billing:  options.Billing,
		Logger:   options.Logger,
		Revision: options.Revision,
	}

	router := gin.New()
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
		httpx.RateLimit(options.Limiter, options.Logger),
	)

	public.GET("/plans", handlers.ListPlans)
	public.GET("/identity/ping", handlers.IdentityPing)

	authenticated := router.Group("/v1")
	authenticated.Use(
		httpx.Authenticate(options.Verifier, options.Denylist, options.Logger),
		httpx.EndpointClass("authenticated_read"),
		httpx.RateLimit(options.Limiter, options.Logger),
		httpx.RequireIdempotencyKey(),
	)

	authenticated.GET("/me", handlers.Me)
	authenticated.POST("/organizations", handlers.CreateOrganization)

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
