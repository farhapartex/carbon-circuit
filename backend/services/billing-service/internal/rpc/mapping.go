package rpc

import (
	billingv1 "github.com/carboncircuit/backend/gen/carboncircuit/billing/v1"
	"github.com/carboncircuit/backend/services/billing-service/internal/domain"
)

var tierToProto = map[domain.PlanTier]billingv1.PlanTier{
	domain.TierBuyer:      billingv1.PlanTier_PLAN_TIER_BUYER,
	domain.TierStarter:    billingv1.PlanTier_PLAN_TIER_STARTER,
	domain.TierGrowth:     billingv1.PlanTier_PLAN_TIER_GROWTH,
	domain.TierEnterprise: billingv1.PlanTier_PLAN_TIER_ENTERPRISE,
}

var organizationTypeFromProto = map[billingv1.OrganizationType]domain.OrganizationType{
	billingv1.OrganizationType_ORGANIZATION_TYPE_MANUFACTURER: domain.OrganizationManufacturer,
	billingv1.OrganizationType_ORGANIZATION_TYPE_ASSEMBLER:    domain.OrganizationAssembler,
	billingv1.OrganizationType_ORGANIZATION_TYPE_LOGISTICS:    domain.OrganizationLogistics,
	billingv1.OrganizationType_ORGANIZATION_TYPE_CREDIT_BUYER: domain.OrganizationCreditBuyer,
}

var organizationTypeToProto = map[domain.OrganizationType]billingv1.OrganizationType{
	domain.OrganizationManufacturer: billingv1.OrganizationType_ORGANIZATION_TYPE_MANUFACTURER,
	domain.OrganizationAssembler:    billingv1.OrganizationType_ORGANIZATION_TYPE_ASSEMBLER,
	domain.OrganizationLogistics:    billingv1.OrganizationType_ORGANIZATION_TYPE_LOGISTICS,
	domain.OrganizationCreditBuyer:  billingv1.OrganizationType_ORGANIZATION_TYPE_CREDIT_BUYER,
}

var dimensionToProto = map[domain.UsageDimension]billingv1.UsageDimension{
	domain.DimensionBatches:     billingv1.UsageDimension_USAGE_DIMENSION_BATCHES,
	domain.DimensionCheckpoints: billingv1.UsageDimension_USAGE_DIMENSION_CHECKPOINTS,
	domain.DimensionFacilities:  billingv1.UsageDimension_USAGE_DIMENSION_FACILITIES,
	domain.DimensionUsers:       billingv1.UsageDimension_USAGE_DIMENSION_USERS,
	domain.DimensionAIReviews:   billingv1.UsageDimension_USAGE_DIMENSION_AI_REVIEWS,
	domain.DimensionStorageGB:   billingv1.UsageDimension_USAGE_DIMENSION_STORAGE_GB,
}

func planToProto(plan domain.Plan) *billingv1.Plan {
	allowed := make([]billingv1.OrganizationType, 0, len(plan.AllowedOrganizationTypes))
	for _, raw := range plan.AllowedOrganizationTypes {
		if mapped, ok := organizationTypeToProto[domain.OrganizationType(raw)]; ok {
			allowed = append(allowed, mapped)
		}
	}

	limits := make([]*billingv1.PlanLimit, 0, len(plan.Limits))
	for _, limit := range plan.Limits {
		limits = append(limits, &billingv1.PlanLimit{
			Dimension:          dimensionToProto[limit.Dimension],
			Included:           limit.Included,
			FairUseCeiling:     limit.FairUseCeiling,
			OverageRateUsd:     valueOr(limit.OverageRateUSD, ""),
			BlocksOnExhaustion: limit.BlocksOnExhaustion,
		})
	}

	return &billingv1.Plan{
		Id:                       plan.ID.String(),
		Tier:                     tierToProto[plan.Tier],
		Name:                     plan.Name,
		Audience:                 plan.Audience,
		MonthlyPriceUsd:          plan.MonthlyPriceUSD,
		PriceNote:                valueOr(plan.PriceNote, ""),
		AllowedOrganizationTypes: allowed,
		EvidenceStorageGb:        plan.EvidenceStorageGB,
		PortalRatePerMinute:      plan.PortalRatePerMinute,
		ApiRatePerMinute:         plan.APIRatePerMinute,
		ApiKeyLimit:              plan.APIKeyLimit,
		MarketplaceFeeBps:        plan.MarketplaceFeeBps,
		ReviewTurnaround:         plan.ReviewTurnaround,
		SupportLevel:             plan.SupportLevel,
		Limits:                   limits,
	}
}

func valueOr[T any](pointer *T, fallback T) T {
	if pointer == nil {
		return fallback
	}
	return *pointer
}
