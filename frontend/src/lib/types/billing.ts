import type { Id, IsoTimestamp } from "@/lib/types/common";
import type { OrganizationType } from "@/lib/types/organization";

export type PlanTier = "buyer" | "starter" | "growth" | "enterprise";

export type PlanLimit = {
  label: string;
  included: number | null;
  fairUseCeiling: number | null;
  overageRateUsd: string | null;
};

export type Plan = {
  tier: PlanTier;
  name: string;
  audience: string;
  monthlyPriceUsd: string | null;
  priceNote: string | null;
  allowedOrganizationTypes: OrganizationType[];
  batchesPerMonth: PlanLimit | null;
  checkpointsPerMonth: PlanLimit | null;
  facilities: PlanLimit | null;
  users: PlanLimit;
  aiReviewedClaimsPerMonth: PlanLimit | null;
  evidenceStorageGb: number | null;
  portalRateLimitPerMinute: number;
  apiRateLimitPerMinute: number | null;
  apiKeyLimit: number | null;
  marketplaceFeeBasisPoints: number | null;
  reviewTurnaround: string;
  supportLevel: string;
};

export type UsageDimension = {
  key: string;
  label: string;
  used: number;
  limit: number | null;
  fairUseCeiling: number | null;
  overageRateUsd: string | null;
  blocksOnExhaustion: boolean;
};

export type PlanUsage = {
  periodStart: IsoTimestamp;
  periodEnd: IsoTimestamp;
  dimensions: UsageDimension[];
};

export type SubscriptionState =
  "active" | "grace_period" | "read_only" | "cancelled";

export type Subscription = {
  planTier: PlanTier;
  state: SubscriptionState;
  gracePeriodEndsAt: IsoTimestamp | null;
  renewsAt: IsoTimestamp | null;
};

export type PaymentMethod = {
  brand: string;
  last4: string;
  expiryMonth: number;
  expiryYear: number;
};

export type Invoice = {
  id: Id;
  number: string;
  amountUsd: string;
  status: "paid" | "open" | "failed" | "void";
  issuedAt: IsoTimestamp;
  paidAt: IsoTimestamp | null;
};

export const PAYMENT_GRACE_PERIOD_DAYS = 14;
