"use server";

import { revalidatePath } from "next/cache";
import { GatewayError } from "@/lib/api/gateway";
import {
  decideClaim,
  type DecisionDraft,
  type ReviewRecord,
} from "@/lib/api/verifier";
import { auth0 } from "@/lib/auth0";

export type DecisionResult =
  { ok: true; review: ReviewRecord } | { ok: false; message: string };

const refusalMessages: Record<string, string> = {
  ABOVE_CEILING:
    "That is above the claim's computed ceiling. Nothing can be issued above it.",
  ABOVE_REQUESTED:
    "That is more than the facility asked for. Approve at or below the requested amount.",
  REASON_TOO_SHORT:
    "A rejection or information request needs a reason of at least 40 characters.",
  ALREADY_DECIDED: "You have already recorded a decision on this claim.",
  APPROVAL_NOT_POSITIVE: "An approved amount must be greater than zero.",
  NOT_IN_REVIEW: "This claim is no longer awaiting a decision.",
};

const codeMessages: Record<string, string> = {
  FORBIDDEN: "Recording a decision requires the verifier role.",
  ORGANIZATION_READ_ONLY: "Recording a decision requires the verifier role.",
  RESOURCE_NOT_FOUND: "That claim is not available for review.",
  IDEMPOTENCY_KEY_REUSED:
    "This decision form was already submitted. Reload the claim.",
  REQUEST_IN_PROGRESS: "This decision is already being recorded.",
  DEPENDENCY_UNAVAILABLE:
    "The review service is unreachable right now. Try again shortly.",
};

const describe = (error: GatewayError): string => {
  const reason = error.codeForField("claim");
  if (reason && refusalMessages[reason]) {
    return refusalMessages[reason];
  }
  return codeMessages[error.code] ?? "The decision could not be recorded.";
};

export const recordDecision = async (
  claimId: string,
  draft: DecisionDraft,
  idempotencyKey: string,
): Promise<DecisionResult> => {
  try {
    const { token } = await auth0.getAccessToken();
    const review = await decideClaim(token, claimId, draft, idempotencyKey);

    revalidatePath("/verifier/queue");
    revalidatePath(`/verifier/queue/${claimId}`);

    return { ok: true, review };
  } catch (error) {
    if (error instanceof GatewayError) {
      return { ok: false, message: describe(error) };
    }
    throw error;
  }
};
