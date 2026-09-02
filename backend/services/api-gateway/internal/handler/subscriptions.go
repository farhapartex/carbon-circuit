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

var planTierByName = map[string]billingv1.PlanTier{
	"buyer":      billingv1.PlanTier_PLAN_TIER_BUYER,
	"starter":    billingv1.PlanTier_PLAN_TIER_STARTER,
	"growth":     billingv1.PlanTier_PLAN_TIER_GROWTH,
	"enterprise": billingv1.PlanTier_PLAN_TIER_ENTERPRISE,
}

var billingOrganizationType = map[identityv1.OrganizationType]billingv1.OrganizationType{
	identityv1.OrganizationType_ORGANIZATION_TYPE_MANUFACTURER: billingv1.OrganizationType_ORGANIZATION_TYPE_MANUFACTURER,
	identityv1.OrganizationType_ORGANIZATION_TYPE_ASSEMBLER:    billingv1.OrganizationType_ORGANIZATION_TYPE_ASSEMBLER,
	identityv1.OrganizationType_ORGANIZATION_TYPE_LOGISTICS:    billingv1.OrganizationType_ORGANIZATION_TYPE_LOGISTICS,
	identityv1.OrganizationType_ORGANIZATION_TYPE_CREDIT_BUYER: billingv1.OrganizationType_ORGANIZATION_TYPE_CREDIT_BUYER,
}

var plansPurchasableBy = map[identityv1.OrganizationRole]bool{
	identityv1.OrganizationRole_ORGANIZATION_ROLE_OWNER: true,
	identityv1.OrganizationRole_ORGANIZATION_ROLE_ADMIN: true,
}

type createSubscriptionRequest struct {
	PlanTier string `json:"plan_tier" binding:"required"`
}

type subscriptionResponse struct {
	PlanTier string `json:"plan_tier"`
	State    string `json:"state"`
}

func (h *Handlers) CreateSubscription(c *gin.Context) {
	caller, verified := auth.CallerFrom(c.Request.Context())
	if !verified {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	var body createSubscriptionRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.Fail(c, httpx.CodeValidation)
		return
	}

	tier, known := planTierByName[body.PlanTier]
	if !known {
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
			Field: "plan_tier",
			Code:  "UNSUPPORTED_VALUE",
		})
		return
	}

	key, present := httpx.IdempotencyKeyFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeIdempotencyKeyRequired)
		return
	}

	resolved, err := h.Identity.ResolveSession(c.Request.Context(), caller)
	if err != nil {
		h.Logger.Error("resolve session upstream failed", errorAttributes(c, err)...)
		httpx.Fail(c, httpx.CodeDependencyUnavailable)
		return
	}

	organization := resolved.GetOrganization()
	if organization == nil {
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
			Field: "organization",
			Code:  "REQUIRED",
		})
		return
	}

	if !plansPurchasableBy[resolved.GetRole()] {
		httpx.Fail(c, httpx.CodeForbidden)
		return
	}

	created, err := h.Billing.CreateSubscription(c.Request.Context(), key,
		&billingv1.CreateSubscriptionRequest{
			OrganizationId:   organization.GetId(),
			OrganizationType: billingOrganizationType[organization.GetType()],
			PlanTier:         tier,
		})
	if err != nil {
		h.failSubscribe(c, err)
		return
	}

	subscription := created.GetSubscription()

	httpx.Data(c, http.StatusCreated, subscriptionResponse{
		PlanTier: tierName[subscription.GetPlanTier()],
		State:    subscriptionStateName[subscription.GetState()],
	})
}

func (h *Handlers) failSubscribe(c *gin.Context, err error) {
	switch status.Code(err) {
	case codes.Aborted:
		httpx.Fail(c, httpx.CodeRequestInProgress)
	case codes.AlreadyExists:
		httpx.Fail(c, httpx.CodeConflict)
	case codes.PermissionDenied:
		httpx.Fail(c, httpx.CodeForbidden)
	case codes.InvalidArgument:
		httpx.Fail(c, httpx.CodeValidation)
	default:
		h.Logger.Error("create subscription upstream failed", errorAttributes(c, err)...)
		httpx.Fail(c, httpx.CodeDependencyUnavailable)
	}
}
