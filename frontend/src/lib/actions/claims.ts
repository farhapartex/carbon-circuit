"use server";

import { randomUUID } from "node:crypto";
import { revalidatePath } from "next/cache";
import {
  previewCeiling,
  submitClaim,
  type CeilingPreview,
  type CeilingRequest,
  type ClaimDraft,
  type ClaimRecord,
} from "@/lib/api/claims";
import { GatewayError } from "@/lib/api/gateway";
import { auth0 } from "@/lib/auth0";

export type ClaimSubmissionResult =
  { ok: true; claim: ClaimRecord } | { ok: false; message: string };

export type CeilingResult =
  { ok: true; preview: CeilingPreview } | { ok: false; message: string };

const refusalMessages: Record<string, string> = {
  ACTIVITY_UNSUPPORTED:
    "This activity type cannot be submitted yet — no operating capacity is recorded for it, so its ceiling cannot be computed.",
  EVIDENCE_REQUIRED:
    "Attach at least one supporting document before submitting.",
  TOO_MUCH_EVIDENCE: "A claim may carry at most 25 supporting documents.",
  EVIDENCE_UNUSABLE:
    "One of the attached documents did not pass scanning. Remove it and attach a replacement.",
  ATTESTATION_REQUIRED: "The exclusivity attestation is required.",
  VINTAGE_CAPACITY_EXHAUSTED:
    "Earlier claims for this facility and vintage have already used the capacity you are asking for. Lower the requested amount, or wait for one of them to be decided.",
  CAPACITY_UNKNOWN:
    "This facility has no recorded operating capacity, so its credit ceiling cannot be computed.",
  PERIOD_OUTSIDE_VINTAGE:
    "The claim period must fall entirely inside its vintage year.",
  PERIOD_EMPTY: "The claim period must cover at least one day.",
  FACTOR_UNKNOWN:
    "No published emission factor covers this facility's grid region for that period.",
  DISCOUNT_UNKNOWN:
    "This facility carries no recognised verification status, so its ceiling cannot be computed.",
};

const codeMessages: Record<string, string> = {
  RESOURCE_NOT_FOUND: "That facility is not registered to this organization.",
  ORGANIZATION_READ_ONLY:
    "This organization cannot submit claims in its current state.",
  RATE_LIMITED:
    "You have reached the claim submission limit for this hour. Try again shortly.",
  IDEMPOTENCY_KEY_REUSED:
    "This submission was already used for a different claim. Reload and try again.",
  REQUEST_IN_PROGRESS: "This claim is already being submitted.",
  DEPENDENCY_UNAVAILABLE:
    "The claims service is unreachable right now. Try again shortly.",
};

const describe = (error: GatewayError): string => {
  const reason = error.codeForField("claim");
  if (reason && refusalMessages[reason]) {
    return refusalMessages[reason];
  }
  return codeMessages[error.code] ?? "The claim could not be submitted.";
};

export const submitSustainabilityClaim = async (
  draft: ClaimDraft,
  idempotencyKey: string,
): Promise<ClaimSubmissionResult> => {
  try {
    const { token } = await auth0.getAccessToken();
    const claim = await submitClaim(token, draft, idempotencyKey);

    revalidatePath("/claims");

    return { ok: true, claim };
  } catch (error) {
    if (error instanceof GatewayError) {
      return { ok: false, message: describe(error) };
    }
    throw error;
  }
};

export const loadCeilingPreview = async (
  request: CeilingRequest,
): Promise<CeilingResult> => {
  try {
    const { token } = await auth0.getAccessToken();
    const preview = await previewCeiling(token, request, randomUUID());

    return { ok: true, preview };
  } catch (error) {
    if (error instanceof GatewayError) {
      return { ok: false, message: describe(error) };
    }
    throw error;
  }
};
