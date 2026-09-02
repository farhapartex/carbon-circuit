package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/carboncircuit/backend/internal/auth"
	"github.com/carboncircuit/backend/internal/httpx"
)

type acceptInvitationRequest struct {
	Token string `json:"token" binding:"required,min=16,max=128"`
}

type acceptedInvitationResponse struct {
	OrganizationID   string `json:"organization_id"`
	OrganizationName string `json:"organization_name"`
	Role             string `json:"role"`
}

func (h *Handlers) AcceptInvitation(c *gin.Context) {
	verified, authenticated := auth.CallerFrom(c.Request.Context())
	if !authenticated {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	var body acceptInvitationRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.Fail(c, httpx.CodeValidation)
		return
	}

	key, present := httpx.IdempotencyKeyFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeIdempotencyKeyRequired)
		return
	}

	accepted, err := h.Identity.AcceptInvitation(c.Request.Context(), key, body.Token)
	if err != nil {
		h.failTeam(c, err)
		return
	}

	h.forgetCaller(c, verified.Subject)

	httpx.Data(c, http.StatusCreated, acceptedInvitationResponse{
		OrganizationID:   accepted.GetOrganizationId(),
		OrganizationName: accepted.GetOrganizationName(),
		Role:             organizationRoleName[accepted.GetRole()],
	})
}
