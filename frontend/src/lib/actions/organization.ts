"use server";

import { GatewayError } from "@/lib/api/gateway";
import { fetchMe } from "@/lib/api/me";
import {
  createOrganization,
  type CreatedOrganization,
  type OrganizationDraft,
} from "@/lib/api/organizations";
import { auth0 } from "@/lib/auth0";
import { TENANCY_KEY, tenancyFrom } from "@/lib/session";

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

  await refreshTenancy(token);

  return { ok: true, organization };
};

const refreshTenancy = async (token: string) => {
  const session = await auth0.getSession();
  if (!session) {
    return;
  }

  try {
    const current = await fetchMe(token);
    await auth0.updateSession({
      ...session,
      [TENANCY_KEY]: tenancyFrom(current),
    });
  } catch {
    return;
  }
};
