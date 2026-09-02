"use server";

import { GatewayError } from "@/lib/api/gateway";
import {
  designateTreasury,
  requestTreasuryChallenge,
  type DesignatedTreasury,
  type TreasuryChallenge,
} from "@/lib/api/treasury";
import { auth0 } from "@/lib/auth0";
import { TENANCY_KEY, tenancyOf, type Tenancy } from "@/lib/session";

export type ChallengeResult =
  { ok: true; challenge: TreasuryChallenge } | { ok: false; code: string };

export type DesignationResult =
  { ok: true; treasury: DesignatedTreasury } | { ok: false; code: string };

export const startTreasuryDesignation = async (
  idempotencyKey: string,
): Promise<ChallengeResult> => {
  const { token } = await auth0.getAccessToken();

  try {
    return {
      ok: true,
      challenge: await requestTreasuryChallenge(token, idempotencyKey),
    };
  } catch (error) {
    if (error instanceof GatewayError) {
      return { ok: false, code: error.code };
    }
    throw error;
  }
};

export const completeTreasuryDesignation = async (
  message: string,
  signature: string,
  idempotencyKey: string,
): Promise<DesignationResult> => {
  const { token } = await auth0.getAccessToken();

  let treasury: DesignatedTreasury;
  try {
    treasury = await designateTreasury(
      token,
      message,
      signature,
      idempotencyKey,
    );
  } catch (error) {
    if (error instanceof GatewayError) {
      return { ok: false, code: error.code };
    }
    throw error;
  }

  await recordDesignated();

  return { ok: true, treasury };
};

const recordDesignated = async () => {
  const session = await auth0.getSession();
  const tenancy = tenancyOf(session);

  if (!session || !tenancy) {
    return;
  }

  const updated: Tenancy = { ...tenancy, isTreasuryDesignated: true };
  await auth0.updateSession({ ...session, [TENANCY_KEY]: updated });
};
