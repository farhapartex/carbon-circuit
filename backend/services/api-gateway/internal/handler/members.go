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

var roleByName = map[string]identityv1.OrganizationRole{
	"owner":  identityv1.OrganizationRole_ORGANIZATION_ROLE_OWNER,
	"admin":  identityv1.OrganizationRole_ORGANIZATION_ROLE_ADMIN,
	"member": identityv1.OrganizationRole_ORGANIZATION_ROLE_MEMBER,
}

var invitationStateName = map[identityv1.InvitationState]string{
	identityv1.InvitationState_INVITATION_STATE_PENDING:  "pending",
	identityv1.InvitationState_INVITATION_STATE_ACCEPTED: "accepted",
	identityv1.InvitationState_INVITATION_STATE_REVOKED:  "revoked",
	identityv1.InvitationState_INVITATION_STATE_EXPIRED:  "expired",
}

type memberResponse struct {
	UserID       string  `json:"user_id"`
	Email        string  `json:"email"`
	Name         string  `json:"name"`
	Role         string  `json:"role"`
	MFAEnrolled  bool    `json:"mfa_enrolled"`
	JoinedAt     *string `json:"joined_at"`
	LastActiveAt *string `json:"last_active_at"`
}

type invitationResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	State         string `json:"state"`
	InvitedAt     string `json:"invited_at"`
	ExpiresAt     string `json:"expires_at"`
	InvitedByName string `json:"invited_by_name"`
}

type teamResponse struct {
	Members     []memberResponse     `json:"members"`
	Invitations []invitationResponse `json:"invitations"`
}

type inviteRequest struct {
	Email string `json:"email" binding:"required,email,max=254"`
	Role  string `json:"role" binding:"required"`
}

type changeRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type issuedInvitationResponse struct {
	Invitation invitationResponse `json:"invitation"`
	Token      string             `json:"token"`
}

func (h *Handlers) ListMembers(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	team, err := h.Identity.ListMembers(c.Request.Context())
	if err != nil {
		h.failTeam(c, err)
		return
	}

	response := teamResponse{
		Members:     make([]memberResponse, 0, len(team.GetMembers())),
		Invitations: make([]invitationResponse, 0, len(team.GetInvitations())),
	}

	for _, member := range team.GetMembers() {
		response.Members = append(response.Members, memberResponse{
			UserID:       member.GetUserId(),
			Email:        member.GetEmail(),
			Name:         member.GetName(),
			Role:         organizationRoleName[member.GetRole()],
			MFAEnrolled:  member.GetMfaEnrolled(),
			JoinedAt:     emptyToNil(member.GetJoinedAt()),
			LastActiveAt: emptyToNil(member.GetLastActiveAt()),
		})
	}

	for _, invitation := range team.GetInvitations() {
		response.Invitations = append(response.Invitations, toInvitationResponse(invitation))
	}

	httpx.Data(c, http.StatusOK, response)
}

func (h *Handlers) InviteMember(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	var body inviteRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.Fail(c, httpx.CodeValidation)
		return
	}

	role, known := roleByName[body.Role]
	if !known {
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{Field: "role", Code: "UNSUPPORTED_VALUE"})
		return
	}

	key, present := httpx.IdempotencyKeyFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeIdempotencyKeyRequired)
		return
	}

	issued, err := h.Identity.InviteMember(c.Request.Context(), key, body.Email, role)
	if err != nil {
		h.failTeam(c, err)
		return
	}

	httpx.Data(c, http.StatusCreated, issuedInvitationResponse{
		Invitation: toInvitationResponse(issued.GetInvitation()),
		Token:      issued.GetToken(),
	})
}

func (h *Handlers) RevokeInvitation(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	key, present := httpx.IdempotencyKeyFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeIdempotencyKeyRequired)
		return
	}

	if err := h.Identity.RevokeInvitation(c.Request.Context(), key, c.Param("invitationId")); err != nil {
		h.failTeam(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handlers) ChangeMemberRole(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	var body changeRoleRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.Fail(c, httpx.CodeValidation)
		return
	}

	role, known := roleByName[body.Role]
	if !known {
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{Field: "role", Code: "UNSUPPORTED_VALUE"})
		return
	}

	key, present := httpx.IdempotencyKeyFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeIdempotencyKeyRequired)
		return
	}

	changed, err := h.Identity.ChangeMemberRole(
		c.Request.Context(), key, c.Param("userId"), role,
	)
	if err != nil {
		h.failTeam(c, err)
		return
	}

	h.forgetCaller(c, changed.GetAffectedSubject())

	httpx.Data(c, http.StatusOK, memberResponse{
		UserID: changed.GetMember().GetUserId(),
		Role:   organizationRoleName[changed.GetMember().GetRole()],
	})
}

func (h *Handlers) RevokeMember(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	key, present := httpx.IdempotencyKeyFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeIdempotencyKeyRequired)
		return
	}

	revoked, err := h.Identity.RevokeMember(c.Request.Context(), key, c.Param("userId"))
	if err != nil {
		h.failTeam(c, err)
		return
	}

	h.forgetCaller(c, revoked.GetAffectedSubject())

	c.Status(http.StatusNoContent)
}

func (h *Handlers) forgetCaller(c *gin.Context, subject string) {
	if subject == "" || h.Resolver == nil {
		return
	}
	h.Resolver.Invalidate(c.Request.Context(), subject)
}

func toInvitationResponse(invitation *identityv1.Invitation) invitationResponse {
	return invitationResponse{
		ID:            invitation.GetId(),
		Email:         invitation.GetEmail(),
		Role:          organizationRoleName[invitation.GetRole()],
		State:         invitationStateName[invitation.GetState()],
		InvitedAt:     invitation.GetInvitedAt(),
		ExpiresAt:     invitation.GetExpiresAt(),
		InvitedByName: invitation.GetInvitedByName(),
	}
}

func (h *Handlers) failTeam(c *gin.Context, err error) {
	switch status.Code(err) {
	case codes.PermissionDenied:
		httpx.Fail(c, httpx.CodeForbidden)
	case codes.FailedPrecondition:
		httpx.Fail(c, httpx.CodeConflict)
	case codes.AlreadyExists:
		httpx.Fail(c, httpx.CodeConflict)
	case codes.NotFound:
		httpx.Fail(c, httpx.CodeResourceNotFound)
	case codes.InvalidArgument:
		httpx.Fail(c, httpx.CodeValidation)
	default:
		h.Logger.Error("team upstream failed", errorAttributes(c, err)...)
		httpx.Fail(c, httpx.CodeDependencyUnavailable)
	}
}
