"use server";

import { GatewayError } from "@/lib/api/gateway";
import {
  createSubscription,
  type CreatedSubscription,
} from "@/lib/api/subscriptions";
import { auth0 } from "@/lib/auth0";
import { TENANCY_KEY, tenancyOf, type Tenancy } from "@/lib/session";
import type { PlanTier } from "@/lib/types/billing";

export type SubmitSubscriptionResult =
  { ok: true; subscription: CreatedSubscription } | { ok: false; code: string };

export const submitSubscription = async (
  planTier: PlanTier,
  idempotencyKey: string,
): Promise<SubmitSubscriptionResult> => {
  const { token } = await auth0.getAccessToken();

  let subscription: CreatedSubscription;
  try {
    subscription = await createSubscription(token, planTier, idempotencyKey);
  } catch (error) {
    if (error instanceof GatewayError) {
      return { ok: false, code: error.code };
    }
    throw error;
  }

  await recordSubscribed(subscription);

  return { ok: true, subscription };
};

const recordSubscribed = async (subscription: CreatedSubscription) => {
  const session = await auth0.getSession();
  const tenancy = tenancyOf(session);

  if (!session || !tenancy) {
    return;
  }

  const updated: Tenancy = {
    ...tenancy,
    isSubscribed: subscription.state !== "cancelled",
  };

  await auth0.updateSession({ ...session, [TENANCY_KEY]: updated });
};
