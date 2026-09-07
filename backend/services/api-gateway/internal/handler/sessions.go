package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/carboncircuit/backend/internal/httpx"
	"github.com/carboncircuit/backend/services/api-gateway/internal/caller"
)

type sessionResponse struct {
	ID         string `json:"id"`
	UserAgent  string `json:"user_agent"`
	IPAddress  string `json:"ip_address"`
	StartedAt  string `json:"started_at"`
	LastSeenAt string `json:"last_seen_at"`
	Current    bool   `json:"current"`
}

func (h *Handlers) ListSessions(c *gin.Context) {
	resolved, present := caller.ContextFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	listed, err := h.Identity.ListSessions(c.Request.Context())
	if err != nil {
		h.failSession(c, err)
		return
	}

	sessions := make([]sessionResponse, 0, len(listed.GetSessions()))
	for _, session := range listed.GetSessions() {
		sessions = append(sessions, sessionResponse{
			ID:         session.GetAuth0SessionId(),
			UserAgent:  session.GetUserAgent(),
			IPAddress:  session.GetIpAddress(),
			StartedAt:  session.GetStartedAt(),
			LastSeenAt: session.GetLastSeenAt(),
			Current:    session.GetAuth0SessionId() == resolved.SessionID,
		})
	}

	httpx.Data(c, http.StatusOK, map[string]any{"sessions": sessions})
}

func (h *Handlers) RevokeSession(c *gin.Context) {
	resolved, present := caller.ContextFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	target := c.Param("sessionId")

	revoked, err := h.Identity.RevokeSession(c.Request.Context(), target)
	if err != nil {
		h.failSession(c, err)
		return
	}

	if h.Denylist != nil {
		if err := h.Denylist.RevokeSession(
			c.Request.Context(), revoked.GetSubject(), target,
		); err != nil {
			h.Logger.Error("session revoked in identity but not denylisted",
				errorAttributes(c, err)...)
			httpx.Fail(c, httpx.CodeInternal)
			return
		}
	}

	if target == resolved.SessionID && h.Resolver != nil {
		h.Resolver.Invalidate(c.Request.Context(), revoked.GetSubject())
	}

	c.Status(http.StatusNoContent)
}

func (h *Handlers) failSession(c *gin.Context, err error) {
	switch status.Code(err) {
	case codes.NotFound:
		httpx.Fail(c, httpx.CodeResourceNotFound)
	case codes.InvalidArgument:
		httpx.Fail(c, httpx.CodeValidation)
	case codes.PermissionDenied:
		httpx.Fail(c, httpx.CodeForbidden)
	default:
		h.Logger.Error("session upstream failed", errorAttributes(c, err)...)
		httpx.Fail(c, httpx.CodeDependencyUnavailable)
	}
}
