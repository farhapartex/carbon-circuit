package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/carboncircuit/backend/internal/httpx"
	"github.com/carboncircuit/backend/services/api-gateway/internal/upstream"
)

type RouterOptions struct {
	Identity    *upstream.Identity
	Logger      *slog.Logger
	Environment string
	Revision    string
}

func NewRouter(options RouterOptions) *gin.Engine {
	if options.Environment != "development" {
		gin.SetMode(gin.ReleaseMode)
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

	v1 := router.Group("/v1")
	v1.GET("/identity/ping", func(c *gin.Context) {
		response, err := options.Identity.Ping(c.Request.Context())
		if err != nil {
			options.Logger.Error("identity ping failed",
				slog.Any("error", err),
				slog.String("request_id", httpx.CorrelationID(c)),
			)
			httpx.Fail(c, httpx.CodeDependencyUnavailable)
			return
		}

		httpx.Data(c, http.StatusOK, gin.H{
			"service":            response.GetService(),
			"revision":           response.GetRevision(),
			"database_reachable": response.GetDatabaseReachable(),
		})
	})

	return router
}
