package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	billingv1 "github.com/carboncircuit/backend/gen/carboncircuit/billing/v1"
	"github.com/carboncircuit/backend/internal/httpx"
	"github.com/carboncircuit/backend/services/api-gateway/internal/caller"
)

var planTierByName = map[string]billingv1.PlanTier{
	"buyer":      billingv1.PlanTier_PLAN_TIER_BUYER,
	"starter":    billingv1.PlanTier_PLAN_TIER_STARTER,
	"growth":     billingv1.PlanTier_PLAN_TIER_GROWTH,
	"enterprise": billingv1.PlanTier_PLAN_TIER_ENTERPRISE,
}

var plansPurchasableBy = map[string]bool{"owner": true, "admin": true}

type createSubscriptionRequest struct {
	PlanTier string `json:"plan_tier" binding:"required"`
}

type subscriptionResponse struct {
	PlanTier string `json:"plan_tier"`
	State    string `json:"state"`
}

func (h *Handlers) CreateSubscription(c *gin.Context) {
	resolved, present := caller.ContextFrom(c.Request.Context())
	if !present {
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

	if !resolved.HasOrganization() {
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
			Field: "organization",
			Code:  "REQUIRED",
		})
		return
	}

	if !plansPurchasableBy[resolved.Role] {
		httpx.Fail(c, httpx.CodeForbidden)
		return
	}

	created, err := h.Billing.CreateSubscription(c.Request.Context(), key,
		&billingv1.CreateSubscriptionRequest{PlanTier: tier})
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
