import "server-only";
import type { SessionData } from "@auth0/nextjs-auth0/types";
import type { CurrentSession } from "@/lib/api/me";
import type { VerificationStatus } from "@/lib/status";
import type {
  OrganizationRole,
  OrganizationState,
  OrganizationType,
} from "@/lib/types/organization";

export const TENANCY_KEY = "tenancy";

export type Tenancy = {
  organizationId: string | null;
  organizationType: OrganizationType | null;
  organizationState: OrganizationState | null;
  verificationStatus: VerificationStatus | null;
  role: OrganizationRole | null;
  isSubscribed: boolean;
  isTreasuryDesignated: boolean;
  isOnboardingDone: boolean;
};

export const tenancyFrom = (session: CurrentSession): Tenancy => ({
  organizationId: session.organization?.id ?? null,
  organizationType: session.organization?.type ?? null,
  organizationState: session.organization?.state ?? null,
  verificationStatus: session.organization?.verificationStatus ?? null,
  role: session.organization?.role ?? null,
  isSubscribed: session.isSubscribed,
  isTreasuryDesignated: session.isTreasuryDesignated,
  isOnboardingDone: session.isOnboardingDone,
});

export const tenancyOf = (session: SessionData | null): Tenancy | null => {
  const stored = session?.[TENANCY_KEY];
  return stored ? (stored as Tenancy) : null;
};

export type OnboardingStep = "organization" | "plan" | "verification" | null;

export const blockingStep = (tenancy: Tenancy): OnboardingStep => {
  if (tenancy.organizationState === "suspended") return "verification";
  if (!tenancy.organizationId) return "organization";
  if (!tenancy.isSubscribed) return "plan";
  return null;
};
