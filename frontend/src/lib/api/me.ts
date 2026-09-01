import "server-only";
import { gatewayGet } from "@/lib/api/gateway";
import type { VerificationStatus } from "@/lib/status";
import type { PlanTier } from "@/lib/types/billing";
import type {
  OrganizationRole,
  OrganizationState,
  OrganizationType,
  PlatformRole,
} from "@/lib/types/organization";

export type SubscriptionState =
  "active" | "grace_period" | "read_only" | "cancelled";

type ApiMe = {
  user: {
    id: string;
    email: string;
    name: string;
    platform_role: PlatformRole | null;
    mfa_enrolled: boolean;
  };
  organization: {
    id: string;
    name: string;
    type: OrganizationType;
    state: OrganizationState;
    verification_status: VerificationStatus;
    role: OrganizationRole;
  } | null;
  subscription: {
    plan_tier: PlanTier;
    state: SubscriptionState;
  } | null;
  is_subscribed: boolean;
  is_treasury_designated: boolean;
  is_onboarding_done: boolean;
};

export type CurrentUser = {
  id: string;
  email: string;
  name: string;
  platformRole: PlatformRole | null;
  mfaEnrolled: boolean;
};

export type CurrentOrganization = {
  id: string;
  name: string;
  type: OrganizationType;
  state: OrganizationState;
  verificationStatus: VerificationStatus;
  role: OrganizationRole;
};

export type CurrentSubscription = {
  planTier: PlanTier;
  state: SubscriptionState;
};

export type CurrentSession = {
  user: CurrentUser;
  organization: CurrentOrganization | null;
  subscription: CurrentSubscription | null;
  isSubscribed: boolean;
  isTreasuryDesignated: boolean;
  isOnboardingDone: boolean;
};

export const fetchMe = async (token: string): Promise<CurrentSession> => {
  const me = await gatewayGet<ApiMe>("/v1/me", token);

  return {
    user: {
      id: me.user.id,
      email: me.user.email,
      name: me.user.name,
      platformRole: me.user.platform_role,
      mfaEnrolled: me.user.mfa_enrolled,
    },
    organization: me.organization
      ? {
          id: me.organization.id,
          name: me.organization.name,
          type: me.organization.type,
          state: me.organization.state,
          verificationStatus: me.organization.verification_status,
          role: me.organization.role,
        }
      : null,
    subscription: me.subscription
      ? { planTier: me.subscription.plan_tier, state: me.subscription.state }
      : null,
    isSubscribed: me.is_subscribed,
    isTreasuryDesignated: me.is_treasury_designated,
    isOnboardingDone: me.is_onboarding_done,
  };
};
