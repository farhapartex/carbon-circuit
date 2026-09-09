import "server-only";
import { gatewayGet, gatewayPost } from "@/lib/api/gateway";
import type { ClaimStatus } from "@/lib/status";
import type { ActivityType } from "@/lib/types";

type ApiClaimStatus =
  | "submitted"
  | "ai_review"
  | "human_review"
  | "approved"
  | "rejected"
  | "more_information_requested";

const claimStatusNames: Record<ApiClaimStatus, ClaimStatus> = {
  submitted: "submitted",
  ai_review: "under_ai_review",
  human_review: "under_human_review",
  approved: "approved",
  rejected: "rejected",
  more_information_requested: "more_information_requested",
};

const apiClaimStatuses: Partial<Record<ClaimStatus, ApiClaimStatus>> = {
  submitted: "submitted",
  under_ai_review: "ai_review",
  under_human_review: "human_review",
  approved: "approved",
  rejected: "rejected",
  more_information_requested: "more_information_requested",
};

export type ClaimPriority = "normal" | "high" | "critical";

type ApiClaimEvidence = {
  evidence_id: string;
  file_name: string;
  media_type: string;
  content_hash: string;
  page_count: number | null;
  byte_size: number;
};

type ApiClaim = {
  id: string;
  facility_id: string;
  facility_name: string;
  activity_type: ActivityType;
  vintage_year: number;
  period_start: string;
  period_end: string;
  declared_figures: Record<string, string>;
  requested_amount: string;
  computed_ceiling: string;
  capacity_basis: string;
  capacity_source: string;
  discount_factor: string;
  reference_factor_value: string;
  reference_factor_id: string;
  reference_lookup_key: string;
  status: ApiClaimStatus;
  priority: ClaimPriority;
  requires_dual_approval: boolean;
  exclusivity_attested_at: string;
  issued_amount: string | null;
  created_at: string;
  evidence: ApiClaimEvidence[];
};

export type ClaimEvidenceRecord = {
  evidenceId: string;
  fileName: string;
  mediaType: string;
  contentHash: string;
  pageCount: number | null;
  byteSize: number;
};

export type ClaimRecord = {
  id: string;
  facilityId: string;
  facilityName: string;
  activityType: ActivityType;
  vintageYear: number;
  periodStart: string;
  periodEnd: string;
  declaredFigures: Record<string, string>;
  requestedAmount: string;
  computedCeiling: string;
  capacityBasis: string;
  capacitySource: string;
  discountFactor: string;
  referenceFactorValue: string;
  referenceFactorId: string;
  referenceLookupKey: string;
  status: ClaimStatus;
  priority: ClaimPriority;
  requiresDualApproval: boolean;
  exclusivityAttestedAt: string;
  issuedAmount: string | null;
  createdAt: string;
  evidence: ClaimEvidenceRecord[];
};

export type ClaimPage = {
  claims: ClaimRecord[];
  nextCursor: string | null;
};

export type CeilingPreview = {
  ceiling: string;
  vintageCeiling: string;
  consumed: string;
  remaining: string;
  periodCeiling: string;
  capacityBasis: string;
  capacitySource: string;
  discountFactor: string;
  referenceValue: string;
  gridRegion: string;
  periodDays: number;
  vintageDays: number;
};

export type ClaimDraft = {
  facilityId: string;
  activityType: ActivityType;
  vintageYear: number;
  periodStart: string;
  periodEnd: string;
  declaredFigures: Record<string, string>;
  requestedAmount: string;
  evidenceIds: string[];
  exclusivityAttested: boolean;
};

export type CeilingRequest = {
  facilityId: string;
  activityType: ActivityType;
  vintageYear: number;
  periodStart: string;
  periodEnd: string;
};

const toEvidence = (attachment: ApiClaimEvidence): ClaimEvidenceRecord => ({
  evidenceId: attachment.evidence_id,
  fileName: attachment.file_name,
  mediaType: attachment.media_type,
  contentHash: attachment.content_hash,
  pageCount: attachment.page_count,
  byteSize: attachment.byte_size,
});

const toClaim = (claim: ApiClaim): ClaimRecord => ({
  id: claim.id,
  facilityId: claim.facility_id,
  facilityName: claim.facility_name,
  activityType: claim.activity_type,
  vintageYear: claim.vintage_year,
  periodStart: claim.period_start,
  periodEnd: claim.period_end,
  declaredFigures: claim.declared_figures ?? {},
  requestedAmount: claim.requested_amount,
  computedCeiling: claim.computed_ceiling,
  capacityBasis: claim.capacity_basis,
  capacitySource: claim.capacity_source,
  discountFactor: claim.discount_factor,
  referenceFactorValue: claim.reference_factor_value,
  referenceFactorId: claim.reference_factor_id,
  referenceLookupKey: claim.reference_lookup_key,
  status: claimStatusNames[claim.status],
  priority: claim.priority,
  requiresDualApproval: claim.requires_dual_approval,
  exclusivityAttestedAt: claim.exclusivity_attested_at,
  issuedAmount: claim.issued_amount,
  createdAt: claim.created_at,
  evidence: (claim.evidence ?? []).map(toEvidence),
});

export const fetchClaims = async (
  token: string,
  status?: ClaimStatus,
): Promise<ClaimPage> => {
  const wanted = status ? apiClaimStatuses[status] : undefined;
  const query = wanted ? `?status=${wanted}` : "";
  const listed = await gatewayGet<{
    claims: ApiClaim[];
    next_cursor: string | null;
  }>(`/v1/claims${query}`, token);

  return {
    claims: listed.claims.map(toClaim),
    nextCursor: listed.next_cursor,
  };
};

export const fetchClaim = async (
  token: string,
  claimId: string,
): Promise<ClaimRecord> => {
  const found = await gatewayGet<ApiClaim>(`/v1/claims/${claimId}`, token);
  return toClaim(found);
};

export const submitClaim = async (
  token: string,
  draft: ClaimDraft,
  idempotencyKey: string,
): Promise<ClaimRecord> => {
  const created = await gatewayPost<ApiClaim>(
    "/v1/claims",
    token,
    {
      facility_id: draft.facilityId,
      activity_type: draft.activityType,
      vintage_year: draft.vintageYear,
      period_start: draft.periodStart,
      period_end: draft.periodEnd,
      declared_figures: draft.declaredFigures,
      requested_amount: draft.requestedAmount,
      evidence_ids: draft.evidenceIds,
      exclusivity_attested: draft.exclusivityAttested,
    },
    idempotencyKey,
  );

  return toClaim(created);
};

export const previewCeiling = async (
  token: string,
  request: CeilingRequest,
  idempotencyKey: string,
): Promise<CeilingPreview> => {
  const preview = await gatewayPost<{
    ceiling: string;
    vintage_ceiling: string;
    consumed: string;
    remaining: string;
    period_ceiling: string;
    capacity_basis: string;
    capacity_source: string;
    discount_factor: string;
    reference_value: string;
    grid_region: string;
    period_days: number;
    vintage_days: number;
  }>(
    "/v1/claims/ceiling-preview",
    token,
    {
      facility_id: request.facilityId,
      activity_type: request.activityType,
      vintage_year: request.vintageYear,
      period_start: request.periodStart,
      period_end: request.periodEnd,
    },
    idempotencyKey,
  );

  return {
    ceiling: preview.ceiling,
    vintageCeiling: preview.vintage_ceiling,
    consumed: preview.consumed,
    remaining: preview.remaining,
    periodCeiling: preview.period_ceiling,
    capacityBasis: preview.capacity_basis,
    capacitySource: preview.capacity_source,
    discountFactor: preview.discount_factor,
    referenceValue: preview.reference_value,
    gridRegion: preview.grid_region,
    periodDays: preview.period_days,
    vintageDays: preview.vintage_days,
  };
};
