"use server";

import { revalidatePath } from "next/cache";
import { GatewayError } from "@/lib/api/gateway";
import { revokeSession } from "@/lib/api/sessions";
import { auth0 } from "@/lib/auth0";

export type SessionActionResult = { ok: true } | { ok: false; code: string };

export const revokeSessionAction = async (
  sessionId: string,
  idempotencyKey: string,
): Promise<SessionActionResult> => {
  try {
    const { token } = await auth0.getAccessToken();
    await revokeSession(token, sessionId, idempotencyKey);

    revalidatePath("/settings/profile");

    return { ok: true };
  } catch (error) {
    if (error instanceof GatewayError) {
      return { ok: false, code: error.code };
    }
    throw error;
  }
};
