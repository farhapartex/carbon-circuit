import type { CreditAmount } from "@/lib/decimal";
import type { ClaimStatus, QueuePriority } from "@/lib/status";
import type {
  ActivityType,
  GridRegion,
  Id,
  IsoTimestamp,
  RecycledMaterial,
  ShippingMethod,
} from "@/lib/types/common";

export type EvidenceScanStatus = "pending" | "clean" | "failed";

export type Evidence = {
  id: Id;
  fileName: string;
  mediaType: string;
  byteSize: number;
  pageCount: number | null;
  contentHash: string;
  scanStatus: EvidenceScanStatus;
  uploadedAt: IsoTimestamp;
};

export type RenewableEnergyFigures = {
  activityType: "renewable_energy";
  verifiedKwh: string;
  gridRegion: GridRegion;
};

export type ReducedEmissionLogisticsFigures = {
  activityType: "reduced_emission_logistics";
  tonneKilometres: string;
  shippingMethod: ShippingMethod;
  actualFactorKgPerTonneKm: string;
};

export type ResponsibleSourcingFigures = {
  activityType: "responsible_sourcing";
  material: RecycledMaterial;
  verifiedQuantity: string;
  quantityUnit: "tonne" | "kilogram";
};

export type ClaimFigures =
  | RenewableEnergyFigures
  | ReducedEmissionLogisticsFigures
  | ResponsibleSourcingFigures;

export type CeilingComputation = {
  ceiling: CreditAmount;
  discountFactor: "1.00" | "0.75" | "0.50";
  discountReason: string;
  referenceFactorValue: string;
  referenceFactorUnit: string;
  referenceTableVersion: string;
};

export type ExtractedFigure = {
  label: string;
  value: string;
  unit: string;
  evidenceId: Id;
  pageNumber: number;
  quotation: string;
};

export type AIReviewVerdict =
  | "corroborated"
  | "corroborated_with_discrepancy"
  | "not_corroborated"
  | "escalated";

export type AIReviewFlag = {
  code: string;
  label: string;
  detail: string;
};

export type AIReviewResult = {
  verdict: AIReviewVerdict;
  confidence: number;
  extractedFigures: ExtractedFigure[];
  discrepancySummary: string | null;
  flags: AIReviewFlag[];
  injectionDetected: boolean;
  completedAt: IsoTimestamp;
};

export type AIReviewAvailability =
  | { state: "available"; result: AIReviewResult }
  | { state: "pending" }
  | { state: "unavailable"; reason: string };

export type VerifierDecisionOutcome =
  "approved" | "rejected" | "more_information_requested";

export type VerifierDecision = {
  id: Id;
  verifierName: string;
  outcome: VerifierDecisionOutcome;
  approvedAmount: CreditAmount | null;
  reason: string;
  decidedAt: IsoTimestamp;
};

export type SustainabilityClaim = {
  id: Id;
  organizationId: Id;
  facilityId: Id;
  facilityName: string;
  activityType: ActivityType;
  figures: ClaimFigures;
  vintageYear: number;
  periodStart: IsoTimestamp;
  periodEnd: IsoTimestamp;
  requestedAmount: CreditAmount;
  ceiling: CeilingComputation;
  status: ClaimStatus;
  priority: QueuePriority;
  requiresDualApproval: boolean;
  evidence: Evidence[];
  decisions: VerifierDecision[];
  issuedAmount: CreditAmount | null;
  exclusivityAttestedAt: IsoTimestamp | null;
  submittedAt: IsoTimestamp;
};

export type VerifierQueueEntry = {
  claimId: Id;
  facilityName: string;
  organizationName: string;
  activityType: ActivityType;
  requestedAmount: CreditAmount;
  ceiling: CreditAmount;
  priority: QueuePriority;
  requiresDualApproval: boolean;
  existingApprovalCount: number;
  submittedAt: IsoTimestamp;
};

export const DUAL_APPROVAL_THRESHOLD_TCO2E = "5000.000000";
export const REJECTION_REASON_MINIMUM_LENGTH = 40;
export const MAXIMUM_EVIDENCE_DOCUMENTS = 25;
export const MAXIMUM_EVIDENCE_BYTES = 25 * 1024 * 1024;
export const MAXIMUM_EVIDENCE_PAGES = 100;
export const ACCEPTED_EVIDENCE_MEDIA_TYPES = [
  "application/pdf",
  "image/png",
  "image/jpeg",
  "text/csv",
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
] as const;
