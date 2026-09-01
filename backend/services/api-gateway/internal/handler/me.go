package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	billingv1 "github.com/carboncircuit/backend/gen/carboncircuit/billing/v1"
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

var subscriptionStateName = map[billingv1.SubscriptionState]string{
	billingv1.SubscriptionState_SUBSCRIPTION_STATE_ACTIVE:       "active",
	billingv1.SubscriptionState_SUBSCRIPTION_STATE_GRACE_PERIOD: "grace_period",
	billingv1.SubscriptionState_SUBSCRIPTION_STATE_READ_ONLY:    "read_only",
	billingv1.SubscriptionState_SUBSCRIPTION_STATE_CANCELLED:    "cancelled",
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
	Role               string `json:"role"`
}

type sessionSubscriptionResponse struct {
	PlanTier string `json:"plan_tier"`
	State    string `json:"state"`
}

type meResponse struct {
	User                 sessionUserResponse          `json:"user"`
	Organization         *sessionOrganizationResponse `json:"organization"`
	Subscription         *sessionSubscriptionResponse `json:"subscription"`
	IsSubscribed         bool                         `json:"is_subscribed"`
	IsTreasuryDesignated bool                         `json:"is_treasury_designated"`
	IsOnboardingDone     bool                         `json:"is_onboarding_done"`
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

	response := toMeResponse(resolved)

	if organization := resolved.GetOrganization(); organization != nil {
		subscription, subscriptionErr := h.Billing.GetSubscription(
			c.Request.Context(), organization.GetId(),
		)
		if subscriptionErr != nil {
			h.Logger.Error("get subscription upstream failed", errorAttributes(c, subscriptionErr)...)
			httpx.Fail(c, httpx.CodeDependencyUnavailable)
			return
		}
		applySubscription(&response, subscription.GetSubscription())
	}

	response.IsOnboardingDone = onboardingComplete(response)

	httpx.Data(c, http.StatusOK, response)
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
	}

	organization := resolved.GetOrganization()
	if organization == nil {
		return response
	}

	response.Organization = &sessionOrganizationResponse{
		ID:                 organization.GetId(),
		Name:               organization.GetName(),
		Type:               identityOrganizationTypeName[organization.GetType()],
		State:              organizationStateName[organization.GetState()],
		VerificationStatus: verificationStatusName[organization.GetVerificationStatus()],
		Role:               organizationRoleName[resolved.GetRole()],
	}
	response.IsTreasuryDesignated = organization.GetTreasuryDesignated()

	return response
}

func applySubscription(response *meResponse, subscription *billingv1.Subscription) {
	if subscription == nil {
		return
	}

	response.Subscription = &sessionSubscriptionResponse{
		PlanTier: tierName[subscription.GetPlanTier()],
		State:    subscriptionStateName[subscription.GetState()],
	}

	response.IsSubscribed =
		subscription.GetState() != billingv1.SubscriptionState_SUBSCRIPTION_STATE_CANCELLED &&
			subscription.GetState() != billingv1.SubscriptionState_SUBSCRIPTION_STATE_UNSPECIFIED
}

func onboardingComplete(response meResponse) bool {
	return response.Organization != nil &&
		response.IsSubscribed &&
		response.Organization.VerificationStatus == "verified" &&
		response.IsTreasuryDesignated
}
