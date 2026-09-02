"use server";

import { GatewayError } from "@/lib/api/gateway";
import {
  createOrganization,
  type CreatedOrganization,
  type OrganizationDraft,
} from "@/lib/api/organizations";
import { auth0 } from "@/lib/auth0";
import { TENANCY_KEY, type Tenancy } from "@/lib/session";

export type SubmitOrganizationResult =
  { ok: true; organization: CreatedOrganization } | { ok: false; code: string };

export const submitOrganization = async (
  draft: OrganizationDraft,
  idempotencyKey: string,
): Promise<SubmitOrganizationResult> => {
  const { token } = await auth0.getAccessToken();

  let organization: CreatedOrganization;
  try {
    organization = await createOrganization(token, draft, idempotencyKey);
  } catch (error) {
    if (error instanceof GatewayError) {
      return { ok: false, code: error.code };
    }
    throw error;
  }

  await recordTenancy(draft, organization);

  return { ok: true, organization };
};

const recordTenancy = async (
  draft: OrganizationDraft,
  organization: CreatedOrganization,
) => {
  const session = await auth0.getSession();
  if (!session) {
    return;
  }

  const tenancy: Tenancy = {
    organizationId: organization.id,
    organizationName: organization.name,
    organizationType: draft.type,
    organizationState: organization.state,
    verificationStatus: organization.verificationStatus,
    role: "owner",
    isSubscribed: false,
    isTreasuryDesignated: false,
    isOnboardingDone: false,
  };

  await auth0.updateSession({ ...session, [TENANCY_KEY]: tenancy });
};
