package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/internal/auth"
	"github.com/carboncircuit/backend/internal/httpx"
)

var identityOrganizationTypeName = map[identityv1.OrganizationType]string{
	identityv1.OrganizationType_ORGANIZATION_TYPE_MANUFACTURER: "manufacturer",
	identityv1.OrganizationType_ORGANIZATION_TYPE_ASSEMBLER:    "assembler",
	identityv1.OrganizationType_ORGANIZATION_TYPE_LOGISTICS:    "logistics",
	identityv1.OrganizationType_ORGANIZATION_TYPE_CREDIT_BUYER: "credit_buyer",
}

var organizationStateName = map[identityv1.OrganizationState]string{
	identityv1.OrganizationState_ORGANIZATION_STATE_ACTIVE:     "active",
	identityv1.OrganizationState_ORGANIZATION_STATE_RESTRICTED: "restricted",
	identityv1.OrganizationState_ORGANIZATION_STATE_READ_ONLY:  "read_only",
	identityv1.OrganizationState_ORGANIZATION_STATE_SUSPENDED:  "suspended",
}

var verificationStatusName = map[identityv1.VerificationStatus]string{
	identityv1.VerificationStatus_VERIFICATION_STATUS_VERIFIED:   "verified",
	identityv1.VerificationStatus_VERIFICATION_STATUS_UNVERIFIED: "unverified",
	identityv1.VerificationStatus_VERIFICATION_STATUS_REJECTED:   "rejected",
}

var organizationRoleName = map[identityv1.OrganizationRole]string{
	identityv1.OrganizationRole_ORGANIZATION_ROLE_OWNER:  "owner",
	identityv1.OrganizationRole_ORGANIZATION_ROLE_ADMIN:  "admin",
	identityv1.OrganizationRole_ORGANIZATION_ROLE_MEMBER: "member",
}

var platformRoleName = map[identityv1.PlatformRole]string{
	identityv1.PlatformRole_PLATFORM_ROLE_VERIFIER:       "verifier",
	identityv1.PlatformRole_PLATFORM_ROLE_PLATFORM_ADMIN: "platform_admin",
}

type sessionUserResponse struct {
	ID           string  `json:"id"`
	Email        string  `json:"email"`
	Name         string  `json:"name"`
	PlatformRole *string `json:"platform_role"`
	MFAEnrolled  bool    `json:"mfa_enrolled"`
}

type sessionOrganizationResponse struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Type               string `json:"type"`
	State              string `json:"state"`
	VerificationStatus string `json:"verification_status"`
}

type meResponse struct {
	User            sessionUserResponse          `json:"user"`
	NeedsOnboarding bool                         `json:"needs_onboarding"`
	Organization    *sessionOrganizationResponse `json:"organization"`
	Role            *string                      `json:"role"`
}

func (h *Handlers) Me(c *gin.Context) {
	caller, verified := auth.CallerFrom(c.Request.Context())
	if !verified {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	if caller.Email == "" {
		h.Logger.Error("token carries no email claim", errorAttributes(c, nil)...)
		httpx.Fail(c, httpx.CodeInternal)
		return
	}

	resolved, err := h.Identity.ResolveSession(c.Request.Context(), caller)
	if err != nil {
		if status.Code(err) == codes.FailedPrecondition {
			httpx.Fail(c, httpx.CodeConflict)
			return
		}
		h.Logger.Error("resolve session upstream failed", errorAttributes(c, err)...)
		httpx.Fail(c, httpx.CodeDependencyUnavailable)
		return
	}

	httpx.Data(c, http.StatusOK, toMeResponse(resolved))
}

func toMeResponse(resolved *identityv1.ResolveSessionResponse) meResponse {
	user := resolved.GetUser()

	response := meResponse{
		User: sessionUserResponse{
			ID:           user.GetId(),
			Email:        user.GetEmail(),
			Name:         user.GetName(),
			PlatformRole: emptyToNil(platformRoleName[user.GetPlatformRole()]),
			MFAEnrolled:  user.GetMfaEnrolled(),
		},
		NeedsOnboarding: resolved.GetNeedsOnboarding(),
	}

	if organization := resolved.GetOrganization(); organization != nil {
		response.Organization = &sessionOrganizationResponse{
			ID:                 organization.GetId(),
			Name:               organization.GetName(),
			Type:               identityOrganizationTypeName[organization.GetType()],
			State:              organizationStateName[organization.GetState()],
			VerificationStatus: verificationStatusName[organization.GetVerificationStatus()],
		}
		response.Role = emptyToNil(organizationRoleName[resolved.GetRole()])
	}

	return response
}
