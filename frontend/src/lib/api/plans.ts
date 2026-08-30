import "server-only";
import { serverConfig } from "@/lib/config/server";
import type { OrganizationType, Plan, PlanLimit, PlanTier } from "@/lib/types";

type ApiPlanLimit = {
  dimension: string;
  included: number | null;
  fair_use_ceiling: number | null;
  overage_rate_usd: string | null;
  blocks_on_exhaustion: boolean;
};

type ApiPlan = {
  id: string;
  tier: PlanTier;
  name: string;
  audience: string;
  monthly_price_usd: string;
  price_note: string | null;
  allowed_organization_types: OrganizationType[];
  evidence_storage_gb: number | null;
  portal_rate_per_minute: number;
  api_rate_per_minute: number | null;
  api_key_limit: number | null;
  marketplace_fee_bps: number | null;
  review_turnaround: string;
  support_level: string;
  limits: ApiPlanLimit[];
};

const DIMENSION_LABELS: Record<string, string> = {
  batches: "Batches per month",
  checkpoints: "Checkpoints per month",
  facilities: "Facilities",
  users: "Users per organization",
  ai_reviews: "AI-reviewed claims per month",
  storage_gb: "Evidence storage",
};

const toPlanLimit = (limit: ApiPlanLimit): PlanLimit => ({
  label: DIMENSION_LABELS[limit.dimension] ?? limit.dimension,
  included: limit.included,
  fairUseCeiling: limit.fair_use_ceiling,
  overageRateUsd: limit.overage_rate_usd,
});

const limitFor = (plan: ApiPlan, dimension: string): PlanLimit | null => {
  const found = plan.limits.find((limit) => limit.dimension === dimension);
  return found ? toPlanLimit(found) : null;
};

const toPlan = (plan: ApiPlan): Plan => ({
  tier: plan.tier,
  name: plan.name,
  audience: plan.audience,
  monthlyPriceUsd: plan.monthly_price_usd.replace(/\.00$/, ""),
  priceNote: plan.price_note,
  allowedOrganizationTypes: plan.allowed_organization_types,
  batchesPerMonth: limitFor(plan, "batches"),
  checkpointsPerMonth: limitFor(plan, "checkpoints"),
  facilities: limitFor(plan, "facilities"),
  users: limitFor(plan, "users") ?? {
    label: "Users per organization",
    included: null,
    fairUseCeiling: null,
    overageRateUsd: null,
  },
  aiReviewedClaimsPerMonth: limitFor(plan, "ai_reviews"),
  evidenceStorageGb: plan.evidence_storage_gb,
  portalRateLimitPerMinute: plan.portal_rate_per_minute,
  apiRateLimitPerMinute: plan.api_rate_per_minute,
  apiKeyLimit: plan.api_key_limit,
  marketplaceFeeBasisPoints: plan.marketplace_fee_bps,
  reviewTurnaround: plan.review_turnaround,
  supportLevel: plan.support_level,
});

export const fetchPlans = async (
  eligibleFor?: OrganizationType,
): Promise<Plan[]> => {
  const url = new URL("/v1/plans", serverConfig.apiGatewayUrl);
  if (eligibleFor) url.searchParams.set("eligible_for", eligibleFor);

  const response = await fetch(url, {
    headers: { Accept: "application/json" },
    next: { revalidate: 300, tags: ["plans"] },
  });

  if (!response.ok) {
    throw new Error(`Plan lookup failed with ${response.status}`);
  }

  const body = (await response.json()) as { data: ApiPlan[] };
  return body.data.map(toPlan);
};
