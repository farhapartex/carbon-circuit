import "server-only";
import { gatewayPost } from "@/lib/api/gateway";
import type { SubscriptionState } from "@/lib/api/me";
import type { PlanTier } from "@/lib/types/billing";

type ApiSubscription = {
  plan_tier: PlanTier;
  state: SubscriptionState;
};

export type CreatedSubscription = {
  planTier: PlanTier;
  state: SubscriptionState;
};

export const createSubscription = async (
  token: string,
  planTier: PlanTier,
  idempotencyKey: string,
): Promise<CreatedSubscription> => {
  const created = await gatewayPost<ApiSubscription>(
    "/v1/subscriptions",
    token,
    { plan_tier: planTier },
    idempotencyKey,
  );

  return { planTier: created.plan_tier, state: created.state };
};
