"use server";

import { revalidatePath } from "next/cache";
import {
  createApiKey,
  revokeApiKey,
  type IssuedApiKey,
} from "@/lib/api/apiKeys";
import { GatewayError } from "@/lib/api/gateway";
import { auth0 } from "@/lib/auth0";

export type CreateApiKeyResult =
  { ok: true; issued: IssuedApiKey } | { ok: false; code: string };

export type ApiKeyActionResult = { ok: true } | { ok: false; code: string };

const failed = (error: unknown): { ok: false; code: string } => {
  if (error instanceof GatewayError) {
    return { ok: false, code: error.code };
  }
  throw error;
};

export const submitApiKey = async (
  name: string,
  idempotencyKey: string,
): Promise<CreateApiKeyResult> => {
  try {
    const { token } = await auth0.getAccessToken();
    const issued = await createApiKey(token, name, idempotencyKey);

    revalidatePath("/settings/api-keys");

    return { ok: true, issued };
  } catch (error) {
    return failed(error);
  }
};

export const revokeApiKeyAction = async (
  keyId: string,
  idempotencyKey: string,
): Promise<ApiKeyActionResult> => {
  try {
    const { token } = await auth0.getAccessToken();
    await revokeApiKey(token, keyId, idempotencyKey);

    revalidatePath("/settings/api-keys");

    return { ok: true };
  } catch (error) {
    return failed(error);
  }
};
