package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/httpx"
)

type RouterOptions struct {
	Database    *gorm.DB
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

	health := NewHealthHandler(options.Database, "identity-service", options.Revision)
	router.GET("/healthz", health.Live)
	router.GET("/readyz", health.Ready)

	v1 := router.Group("/v1")
	v1.GET("/ping", func(c *gin.Context) {
		httpx.Data(c, 200, gin.H{"pong": true})
	})

	return router
}
