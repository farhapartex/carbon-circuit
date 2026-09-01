import "server-only";
import { gatewayGet } from "@/lib/api/gateway";
import type {
  OrganizationRole,
  OrganizationState,
  OrganizationType,
  PlatformRole,
} from "@/lib/types/organization";
import type { VerificationStatus } from "@/lib/status";

type ApiMe = {
  user: {
    id: string;
    email: string;
    name: string;
    platform_role: PlatformRole | null;
    mfa_enrolled: boolean;
  };
  needs_onboarding: boolean;
  organization: {
    id: string;
    name: string;
    type: OrganizationType;
    state: OrganizationState;
    verification_status: VerificationStatus;
  } | null;
  role: OrganizationRole | null;
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
};

export type CurrentSession = {
  user: CurrentUser;
  needsOnboarding: boolean;
  organization: CurrentOrganization | null;
  role: OrganizationRole | null;
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
    needsOnboarding: me.needs_onboarding,
    organization: me.organization
      ? {
          id: me.organization.id,
          name: me.organization.name,
          type: me.organization.type,
          state: me.organization.state,
          verificationStatus: me.organization.verification_status,
        }
      : null,
    role: me.role,
  };
};
