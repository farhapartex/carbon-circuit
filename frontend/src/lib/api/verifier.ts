import "server-only";
import type { ClaimRecord } from "@/lib/api/claims";
import { gatewayGet, gatewayPost } from "@/lib/api/gateway";

export type DecisionOutcome =
  "approved" | "rejected" | "more_information_requested";

type ApiAIReview = {
  assessment: string;
  confidence: string | null;
  extracted_figures: Record<string, string> | null;
  flags: string[] | null;
  narrative: string;
  assessed_by: string;
  assessed_at: string;
};

type ApiDecision = {
  id: string;
  verifier_user_id: string;
  verifier_name: string;
  outcome: DecisionOutcome;
  approved_amount: string | null;
  reason: string;
  decided_at: string;
};

export type AIReviewRecord = {
  assessment: string;
  confidence: string | null;
  extractedFigures: Record<string, string>;
  flags: string[];
  narrative: string;
  assessedBy: string;
  assessedAt: string;
};

export type DecisionRecord = {
  id: string;
  verifierUserId: string;
  verifierName: string;
  outcome: DecisionOutcome;
  approvedAmount: string | null;
  reason: string;
  decidedAt: string;
};

export type ReviewRecord = {
  claim: ClaimRecord;
  aiReview: AIReviewRecord | null;
  decisions: DecisionRecord[];
};

const toAIReview = (review: ApiAIReview | null): AIReviewRecord | null =>
  review
    ? {
        assessment: review.assessment,
        confidence: review.confidence,
        extractedFigures: review.extracted_figures ?? {},
        flags: review.flags ?? [],
        narrative: review.narrative,
        assessedBy: review.assessed_by,
        assessedAt: review.assessed_at,
      }
    : null;

const toDecision = (decision: ApiDecision): DecisionRecord => ({
  id: decision.id,
  verifierUserId: decision.verifier_user_id,
  verifierName: decision.verifier_name,
  outcome: decision.outcome,
  approvedAmount: decision.approved_amount,
  reason: decision.reason,
  decidedAt: decision.decided_at,
});

export const fetchReviewQueue = async (
  token: string,
): Promise<ClaimRecord[]> => {
  const { fetchClaimsFromPayload } = await import("@/lib/api/claims");
  const queued = await gatewayGet<{ claims: unknown[] }>(
    "/v1/verifier/queue",
    token,
  );
  return fetchClaimsFromPayload(queued.claims);
};

export const fetchReview = async (
  token: string,
  claimId: string,
): Promise<ReviewRecord> => {
  const { fetchClaimsFromPayload } = await import("@/lib/api/claims");
  const reviewed = await gatewayGet<{
    claim: unknown;
    ai_review: ApiAIReview | null;
    decisions: ApiDecision[];
  }>(`/v1/verifier/claims/${claimId}`, token);

  return {
    claim: fetchClaimsFromPayload([reviewed.claim])[0]!,
    aiReview: toAIReview(reviewed.ai_review),
    decisions: reviewed.decisions.map(toDecision),
  };
};

export type DecisionDraft = {
  outcome: DecisionOutcome;
  approvedAmount?: string;
  reason?: string;
};

export const decideClaim = async (
  token: string,
  claimId: string,
  draft: DecisionDraft,
  idempotencyKey: string,
): Promise<ReviewRecord> => {
  const { fetchClaimsFromPayload } = await import("@/lib/api/claims");
  const decided = await gatewayPost<{
    claim: unknown;
    decisions: ApiDecision[];
  }>(
    `/v1/verifier/claims/${claimId}/decision`,
    token,
    {
      outcome: draft.outcome,
      approved_amount: draft.approvedAmount ?? "",
      reason: draft.reason ?? "",
    },
    idempotencyKey,
  );

  return {
    claim: fetchClaimsFromPayload([decided.claim])[0]!,
    aiReview: null,
    decisions: decided.decisions.map(toDecision),
  };
};
