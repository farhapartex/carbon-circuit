package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/httpx"
)

type HealthHandler struct {
	database    *gorm.DB
	serviceName string
	revision    string
}

func NewHealthHandler(database *gorm.DB, serviceName, revision string) *HealthHandler {
	return &HealthHandler{database: database, serviceName: serviceName, revision: revision}
}

func (h *HealthHandler) Live(c *gin.Context) {
	httpx.Data(c, http.StatusOK, gin.H{
		"service":  h.serviceName,
		"revision": h.revision,
		"status":   "alive",
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	pool, err := h.database.DB()
	if err != nil {
		httpx.Fail(c, httpx.CodeDependencyUnavailable)
		return
	}

	if err := pool.PingContext(c.Request.Context()); err != nil {
		httpx.Fail(c, httpx.CodeDependencyUnavailable)
		return
	}

	httpx.Data(c, http.StatusOK, gin.H{
		"service":  h.serviceName,
		"status":   "ready",
		"database": "reachable",
	})
}
