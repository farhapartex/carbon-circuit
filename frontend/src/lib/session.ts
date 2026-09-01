import "server-only";
import type { SessionData } from "@auth0/nextjs-auth0/types";
import type { CurrentSession } from "@/lib/api/me";
import type {
  OrganizationRole,
  OrganizationType,
} from "@/lib/types/organization";

export const TENANCY_KEY = "tenancy";

export type Tenancy = {
  needsOnboarding: boolean;
  organizationId: string | null;
  organizationType: OrganizationType | null;
  role: OrganizationRole | null;
};

export const tenancyFrom = (session: CurrentSession): Tenancy => ({
  needsOnboarding: session.needsOnboarding,
  organizationId: session.organization?.id ?? null,
  organizationType: session.organization?.type ?? null,
  role: session.role,
});

export const tenancyOf = (session: SessionData | null): Tenancy | null => {
  const stored = session?.[TENANCY_KEY];
  return stored ? (stored as Tenancy) : null;
};
