"use server";

import { revalidatePath } from "next/cache";
import {
  createFacility,
  type FacilityDraft,
  type FacilityRecord,
} from "@/lib/api/facilities";
import { GatewayError } from "@/lib/api/gateway";
import { auth0 } from "@/lib/auth0";

export type FacilityResult =
  { ok: true; facility: FacilityRecord } | { ok: false; code: string };

export const submitFacility = async (
  draft: FacilityDraft,
  idempotencyKey: string,
): Promise<FacilityResult> => {
  try {
    const { token } = await auth0.getAccessToken();
    const facility = await createFacility(token, draft, idempotencyKey);

    revalidatePath("/facilities");

    return { ok: true, facility };
  } catch (error) {
    if (error instanceof GatewayError) {
      return { ok: false, code: error.code };
    }
    throw error;
  }
};
