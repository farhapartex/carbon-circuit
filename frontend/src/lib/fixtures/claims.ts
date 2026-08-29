import { creditAmount } from "@/lib/decimal";
import {
  facilityMatched,
  organizationMatched,
  selfDeclared,
  verifiedManufacturer,
} from "@/lib/fixtures/organizations";
import type {
  AIReviewAvailability,
  Evidence,
  SustainabilityClaim,
  VerifierQueueEntry,
} from "@/lib/types";

const utilityStatement: Evidence = {
  id: "evd_utility_q2",
  fileName: "taipower-statement-2026-q2.pdf",
  mediaType: "application/pdf",
  byteSize: 2_418_332,
  pageCount: 14,
  contentHash:
    "9f2c4a6e8b0d2f4a6c8e0b2d4f6a8c0e2b4d6f8a0c2e4b6d8f0a2c4e6b8d0f2a",
  scanStatus: "clean",
  uploadedAt: "2026-08-05T09:14:00Z",
};

const meterExport: Evidence = {
  id: "evd_meter_export",
  fileName: "solar-meter-export-2026-q2.csv",
  mediaType: "text/csv",
  byteSize: 884_112,
  pageCount: null,
  contentHash:
    "3d5f7a9c1e3b5d7f9a1c3e5b7d9f1a3c5e7b9d1f3a5c7e9b1d3f5a7c9e1b3d5f",
  scanStatus: "clean",
  uploadedAt: "2026-08-05T09:16:00Z",
};

const auditReport: Evidence = {
  id: "evd_audit_report",
  fileName: "third-party-audit-2026.pdf",
  mediaType: "application/pdf",
  byteSize: 5_120_774,
  pageCount: 38,
  contentHash:
    "7a9c1e3b5d7f9a1c3e5b7d9f1a3c5e7b9d1f3a5c7e9b1d3f5a7c9e1b3d5f7a9c",
  scanStatus: "clean",
  uploadedAt: "2026-08-05T09:19:00Z",
};

export const approvedRenewableClaim: SustainabilityClaim = {
  id: "clm_renewable_2026_q2",
  organizationId: verifiedManufacturer.id,
  facilityId: facilityMatched.id,
  facilityName: facilityMatched.name,
  activityType: "renewable_energy",
  figures: {
    activityType: "renewable_energy",
    verifiedKwh: "12400000",
    gridRegion: "TW",
  },
  vintageYear: 2026,
  periodStart: "2026-04-01T00:00:00Z",
  periodEnd: "2026-06-30T23:59:59Z",
  requestedAmount: creditAmount("6125.600000"),
  ceiling: {
    ceiling: creditAmount("6125.600000"),
    discountFactor: "1.00",
    discountReason:
      "Facility matched in the facility registry dataset, so attested capacity is used.",
    referenceFactorValue: "0.494",
    referenceFactorUnit: "kgCO2e/kWh",
    referenceTableVersion: "grid-factors-2026.01",
  },
  status: "approved",
  priority: "critical",
  requiresDualApproval: true,
  evidence: [utilityStatement, meterExport, auditReport],
  decisions: [
    {
      id: "dec_first",
      verifierName: "Anneke Visser",
      outcome: "approved",
      approvedAmount: creditAmount("6016.920000"),
      reason:
        "Extracted supply figure of 12,180,000 kWh is corroborated by the Taipower statement on page 4 and the meter export. Approving at the extracted figure rather than the declared figure.",
      decidedAt: "2026-08-11T15:42:00Z",
    },
    {
      id: "dec_second",
      verifierName: "Marcus Oyelaran",
      outcome: "approved",
      approvedAmount: creditAmount("6016.920000"),
      reason:
        "Second approval. Reviewed the 1.8% discrepancy independently and agree the extracted figure is the defensible one.",
      decidedAt: "2026-08-12T13:20:00Z",
    },
  ],
  issuedAmount: creditAmount("6016.920000"),
  exclusivityAttestedAt: "2026-08-05T09:25:00Z",
  submittedAt: "2026-08-05T09:26:00Z",
};

export const humanReviewClaim: SustainabilityClaim = {
  id: "clm_logistics_2026_q2",
  organizationId: verifiedManufacturer.id,
  facilityId: organizationMatched.id,
  facilityName: organizationMatched.name,
  activityType: "reduced_emission_logistics",
  figures: {
    activityType: "reduced_emission_logistics",
    tonneKilometres: "4820000",
    shippingMethod: "sea_freight_container",
    actualFactorKgPerTonneKm: "0.011",
  },
  vintageYear: 2026,
  periodStart: "2026-04-01T00:00:00Z",
  periodEnd: "2026-06-30T23:59:59Z",
  requestedAmount: creditAmount("24.100000"),
  ceiling: {
    ceiling: creditAmount("18.075000"),
    discountFactor: "0.75",
    discountReason:
      "Organization is registry-matched but this facility is not individually matched.",
    referenceFactorValue: "0.016",
    referenceFactorUnit: "kgCO2e/tonne-km",
    referenceTableVersion: "logistics-baselines-2026.01",
  },
  status: "under_human_review",
  priority: "high",
  requiresDualApproval: false,
  evidence: [auditReport],
  decisions: [],
  issuedAmount: null,
  exclusivityAttestedAt: "2026-08-24T11:02:00Z",
  submittedAt: "2026-08-24T11:03:00Z",
};

export const aiReviewClaim: SustainabilityClaim = {
  id: "clm_sourcing_2026_q3",
  organizationId: verifiedManufacturer.id,
  facilityId: facilityMatched.id,
  facilityName: facilityMatched.name,
  activityType: "responsible_sourcing",
  figures: {
    activityType: "responsible_sourcing",
    material: "copper",
    verifiedQuantity: "412.5",
    quantityUnit: "tonne",
  },
  vintageYear: 2026,
  periodStart: "2026-07-01T00:00:00Z",
  periodEnd: "2026-09-30T23:59:59Z",
  requestedAmount: creditAmount("1072.500000"),
  ceiling: {
    ceiling: creditAmount("1072.500000"),
    discountFactor: "1.00",
    discountReason:
      "Facility matched in the facility registry dataset, so attested capacity is used.",
    referenceFactorValue: "2.60",
    referenceFactorUnit: "tCO2e/tonne",
    referenceTableVersion: "material-factors-2026.01",
  },
  status: "under_ai_review",
  priority: "normal",
  requiresDualApproval: false,
  evidence: [utilityStatement],
  decisions: [],
  issuedAmount: null,
  exclusivityAttestedAt: "2026-08-28T16:40:00Z",
  submittedAt: "2026-08-28T16:41:00Z",
};

export const submittedClaim: SustainabilityClaim = {
  ...aiReviewClaim,
  id: "clm_submitted",
  status: "submitted",
  priority: "normal",
  submittedAt: "2026-08-29T07:50:00Z",
};

export const rejectedClaim: SustainabilityClaim = {
  id: "clm_rejected",
  organizationId: "org_unverified_assembler",
  facilityId: selfDeclared.id,
  facilityName: selfDeclared.name,
  activityType: "renewable_energy",
  figures: {
    activityType: "renewable_energy",
    verifiedKwh: "9800000",
    gridRegion: "MY",
  },
  vintageYear: 2026,
  periodStart: "2026-01-01T00:00:00Z",
  periodEnd: "2026-03-31T23:59:59Z",
  requestedAmount: creditAmount("5733.000000"),
  ceiling: {
    ceiling: creditAmount("1989.000000"),
    discountFactor: "0.50",
    discountReason:
      "No registry match, so nothing about the declared scale is independently corroborated.",
    referenceFactorValue: "0.585",
    referenceFactorUnit: "kgCO2e/kWh",
    referenceTableVersion: "grid-factors-2026.01",
  },
  status: "rejected",
  priority: "high",
  requiresDualApproval: false,
  evidence: [utilityStatement],
  decisions: [
    {
      id: "dec_reject",
      verifierName: "Anneke Visser",
      outcome: "rejected",
      approvedAmount: null,
      reason:
        "The submitted statement covers a different facility address than the one registered, and the declared consumption exceeds the facility's own declared annual figure for a single quarter.",
      decidedAt: "2026-07-19T10:15:00Z",
    },
  ],
  issuedAmount: null,
  exclusivityAttestedAt: "2026-07-02T08:00:00Z",
  submittedAt: "2026-07-02T08:01:00Z",
};

export const moreInformationClaim: SustainabilityClaim = {
  ...humanReviewClaim,
  id: "clm_more_info",
  status: "more_information_requested",
  decisions: [
    {
      id: "dec_more_info",
      verifierName: "Marcus Oyelaran",
      outcome: "more_information_requested",
      approvedAmount: null,
      reason:
        "The carrier fuel records cover only two of the three months in the claim period. Please supply the missing month before this can be decided.",
      decidedAt: "2026-08-26T09:30:00Z",
    },
  ],
};

export const draftClaim: SustainabilityClaim = {
  ...aiReviewClaim,
  id: "clm_draft",
  status: "draft",
  evidence: [],
  exclusivityAttestedAt: null,
  submittedAt: "2026-08-29T08:10:00Z",
};

export const claims: SustainabilityClaim[] = [
  draftClaim,
  submittedClaim,
  aiReviewClaim,
  humanReviewClaim,
  moreInformationClaim,
  approvedRenewableClaim,
  rejectedClaim,
];

export const aiReviewByClaimId: Record<string, AIReviewAvailability> = {
  [approvedRenewableClaim.id]: {
    state: "available",
    result: {
      verdict: "corroborated_with_discrepancy",
      confidence: 0.87,
      extractedFigures: [
        {
          label: "Renewable supply for period",
          value: "12180000",
          unit: "kWh",
          evidenceId: utilityStatement.id,
          pageNumber: 4,
          quotation:
            "Total renewable supply delivered under contract 44-812 for the period: 12,180,000 kWh",
        },
        {
          label: "Meter-recorded solar generation",
          value: "12179450",
          unit: "kWh",
          evidenceId: meterExport.id,
          pageNumber: 1,
          quotation: "SUM(generation_kwh) = 12179450",
        },
      ],
      discrepancySummary:
        "Declared figure of 12,400,000 kWh exceeds the extracted figure of 12,180,000 kWh by 1.8%. The two independent evidence sources agree with each other to within 0.005%.",
      flags: [
        {
          code: "declared_exceeds_extracted",
          label: "Declared figure above extracted",
          detail: "1.8% above the corroborated figure.",
        },
      ],
      injectionDetected: false,
      completedAt: "2026-08-05T09:34:00Z",
    },
  },
  [humanReviewClaim.id]: {
    state: "unavailable",
    reason:
      "The model provider was unavailable for more than 4 hours, so this claim was promoted to human-only review. No AI assessment is available for it.",
  },
  [aiReviewClaim.id]: { state: "pending" },
};

export const verifierQueue: VerifierQueueEntry[] = [
  {
    claimId: aiReviewClaim.id,
    facilityName: facilityMatched.name,
    organizationName: verifiedManufacturer.name,
    activityType: "responsible_sourcing",
    requestedAmount: creditAmount("1072.500000"),
    ceiling: creditAmount("1072.500000"),
    priority: "normal",
    requiresDualApproval: false,
    existingApprovalCount: 0,
    submittedAt: aiReviewClaim.submittedAt,
  },
  {
    claimId: humanReviewClaim.id,
    facilityName: organizationMatched.name,
    organizationName: verifiedManufacturer.name,
    activityType: "reduced_emission_logistics",
    requestedAmount: creditAmount("24.100000"),
    ceiling: creditAmount("18.075000"),
    priority: "high",
    requiresDualApproval: false,
    existingApprovalCount: 0,
    submittedAt: humanReviewClaim.submittedAt,
  },
  {
    claimId: "clm_awaiting_second",
    facilityName: facilityMatched.name,
    organizationName: verifiedManufacturer.name,
    activityType: "renewable_energy",
    requestedAmount: creditAmount("7420.000000"),
    ceiling: creditAmount("7420.000000"),
    priority: "critical",
    requiresDualApproval: true,
    existingApprovalCount: 1,
    submittedAt: "2026-08-27T12:00:00Z",
  },
];
