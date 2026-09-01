package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	billingv1 "github.com/carboncircuit/backend/gen/carboncircuit/billing/v1"
	"github.com/carboncircuit/backend/internal/httpx"
)

var billingOrganizationTypeByName = map[string]billingv1.OrganizationType{
	"manufacturer": billingv1.OrganizationType_ORGANIZATION_TYPE_MANUFACTURER,
	"assembler":    billingv1.OrganizationType_ORGANIZATION_TYPE_ASSEMBLER,
	"logistics":    billingv1.OrganizationType_ORGANIZATION_TYPE_LOGISTICS,
	"credit_buyer": billingv1.OrganizationType_ORGANIZATION_TYPE_CREDIT_BUYER,
}

var tierName = map[billingv1.PlanTier]string{
	billingv1.PlanTier_PLAN_TIER_BUYER:      "buyer",
	billingv1.PlanTier_PLAN_TIER_STARTER:    "starter",
	billingv1.PlanTier_PLAN_TIER_GROWTH:     "growth",
	billingv1.PlanTier_PLAN_TIER_ENTERPRISE: "enterprise",
}

var organizationTypeName = map[billingv1.OrganizationType]string{
	billingv1.OrganizationType_ORGANIZATION_TYPE_MANUFACTURER: "manufacturer",
	billingv1.OrganizationType_ORGANIZATION_TYPE_ASSEMBLER:    "assembler",
	billingv1.OrganizationType_ORGANIZATION_TYPE_LOGISTICS:    "logistics",
	billingv1.OrganizationType_ORGANIZATION_TYPE_CREDIT_BUYER: "credit_buyer",
}

var dimensionName = map[billingv1.UsageDimension]string{
	billingv1.UsageDimension_USAGE_DIMENSION_BATCHES:     "batches",
	billingv1.UsageDimension_USAGE_DIMENSION_CHECKPOINTS: "checkpoints",
	billingv1.UsageDimension_USAGE_DIMENSION_FACILITIES:  "facilities",
	billingv1.UsageDimension_USAGE_DIMENSION_USERS:       "users",
	billingv1.UsageDimension_USAGE_DIMENSION_AI_REVIEWS:  "ai_reviews",
	billingv1.UsageDimension_USAGE_DIMENSION_STORAGE_GB:  "storage_gb",
}

type planLimitResponse struct {
	Dimension          string  `json:"dimension"`
	Included           *int64  `json:"included"`
	FairUseCeiling     *int64  `json:"fair_use_ceiling"`
	OverageRateUSD     *string `json:"overage_rate_usd"`
	BlocksOnExhaustion bool    `json:"blocks_on_exhaustion"`
}

type planResponse struct {
	ID                       string              `json:"id"`
	Tier                     string              `json:"tier"`
	Name                     string              `json:"name"`
	Audience                 string              `json:"audience"`
	MonthlyPriceUSD          string              `json:"monthly_price_usd"`
	PriceNote                *string             `json:"price_note"`
	AllowedOrganizationTypes []string            `json:"allowed_organization_types"`
	EvidenceStorageGB        *int32              `json:"evidence_storage_gb"`
	PortalRatePerMinute      int32               `json:"portal_rate_per_minute"`
	APIRatePerMinute         *int32              `json:"api_rate_per_minute"`
	APIKeyLimit              *int32              `json:"api_key_limit"`
	MarketplaceFeeBps        *int32              `json:"marketplace_fee_bps"`
	ReviewTurnaround         string              `json:"review_turnaround"`
	SupportLevel             string              `json:"support_level"`
	Limits                   []planLimitResponse `json:"limits"`
}

func (h *Handlers) ListPlans(c *gin.Context) {
	eligibleFor := billingv1.OrganizationType_ORGANIZATION_TYPE_UNSPECIFIED

	if requested := c.Query("eligible_for"); requested != "" {
		mapped, known := billingOrganizationTypeByName[requested]
		if !known {
			httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
				Field: "eligible_for",
				Code:  "UNSUPPORTED_VALUE",
			})
			return
		}
		eligibleFor = mapped
	}

	upstreamResponse, err := h.Billing.ListPlans(c.Request.Context(), eligibleFor)
	if err != nil {
		h.Logger.Error("list plans upstream failed", errorAttributes(c, err)...)
		httpx.Fail(c, httpx.CodeDependencyUnavailable)
		return
	}

	plans := make([]planResponse, 0, len(upstreamResponse.GetPlans()))
	for _, plan := range upstreamResponse.GetPlans() {
		plans = append(plans, toPlanResponse(plan))
	}

	httpx.Data(c, http.StatusOK, plans)
}

func toPlanResponse(plan *billingv1.Plan) planResponse {
	allowed := make([]string, 0, len(plan.GetAllowedOrganizationTypes()))
	for _, organizationType := range plan.GetAllowedOrganizationTypes() {
		allowed = append(allowed, organizationTypeName[organizationType])
	}

	limits := make([]planLimitResponse, 0, len(plan.GetLimits()))
	for _, limit := range plan.GetLimits() {
		limits = append(limits, planLimitResponse{
			Dimension:          dimensionName[limit.GetDimension()],
			Included:           limit.Included,
			FairUseCeiling:     limit.FairUseCeiling,
			OverageRateUSD:     emptyToNil(limit.GetOverageRateUsd()),
			BlocksOnExhaustion: limit.GetBlocksOnExhaustion(),
		})
	}

	return planResponse{
		ID:                       plan.GetId(),
		Tier:                     tierName[plan.GetTier()],
		Name:                     plan.GetName(),
		Audience:                 plan.GetAudience(),
		MonthlyPriceUSD:          plan.GetMonthlyPriceUsd(),
		PriceNote:                emptyToNil(plan.GetPriceNote()),
		AllowedOrganizationTypes: allowed,
		EvidenceStorageGB:        plan.EvidenceStorageGb,
		PortalRatePerMinute:      plan.GetPortalRatePerMinute(),
		APIRatePerMinute:         plan.ApiRatePerMinute,
		APIKeyLimit:              plan.ApiKeyLimit,
		MarketplaceFeeBps:        plan.MarketplaceFeeBps,
		ReviewTurnaround:         plan.GetReviewTurnaround(),
		SupportLevel:             plan.GetSupportLevel(),
		Limits:                   limits,
	}
}

func emptyToNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
