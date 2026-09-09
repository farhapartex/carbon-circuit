import "server-only";
import type { ClaimRecord } from "@/lib/api/claims";
import { creditAmount } from "@/lib/decimal";
import type {
  ClaimFigures,
  Evidence,
  GridRegion,
  RecycledMaterial,
  ShippingMethod,
  SustainabilityClaim,
} from "@/lib/types";

const discountReasons: Record<string, string> = {
  "1.00":
    "This facility matched the facility registry, so its attested capacity is independently sourced and its ceiling carries no reduction.",
  "0.75":
    "The organization matched the registry but this facility did not, so its scale is self-declared and its ceiling is reduced to three quarters.",
  "0.50":
    "Nothing about this facility's declared scale is corroborated, so its ceiling is reduced by half.",
};

const figuresFrom = (claim: ClaimRecord): ClaimFigures => {
  const declared = claim.declaredFigures;

  if (claim.activityType === "reduced_emission_logistics") {
    return {
      activityType: "reduced_emission_logistics",
      tonneKilometres: declared.tonne_kilometres ?? "0",
      shippingMethod: (declared.shipping_method ??
        "road_hgv") as ShippingMethod,
      actualFactorKgPerTonneKm: declared.actual_factor_kg_per_tonne_km ?? "0",
    };
  }

  if (claim.activityType === "responsible_sourcing") {
    return {
      activityType: "responsible_sourcing",
      material: (declared.material ?? "steel") as RecycledMaterial,
      verifiedQuantity: declared.verified_quantity ?? "0",
      quantityUnit:
        declared.quantity_unit === "kilogram" ? "kilogram" : "tonne",
    };
  }

  return {
    activityType: "renewable_energy",
    verifiedKwh: declared.verified_kwh ?? "0",
    gridRegion: (declared.grid_region ?? "TW") as GridRegion,
  };
};

const evidenceFrom = (claim: ClaimRecord): Evidence[] =>
  claim.evidence.map((attachment) => ({
    id: attachment.evidenceId,
    fileName: attachment.fileName,
    mediaType: attachment.mediaType,
    byteSize: attachment.byteSize,
    pageCount: attachment.pageCount,
    contentHash: attachment.contentHash,
    scanStatus: "clean",
    uploadedAt: claim.createdAt,
  }));

export const toSustainabilityClaim = (
  claim: ClaimRecord,
): SustainabilityClaim => ({
  id: claim.id,
  organizationId: "",
  facilityId: claim.facilityId,
  facilityName: claim.facilityName,
  activityType: claim.activityType,
  figures: figuresFrom(claim),
  vintageYear: claim.vintageYear,
  periodStart: claim.periodStart,
  periodEnd: claim.periodEnd,
  requestedAmount: creditAmount(claim.requestedAmount),
  ceiling: {
    ceiling: creditAmount(claim.computedCeiling),
    discountFactor: claim.discountFactor as "1.00" | "0.75" | "0.50",
    discountReason:
      discountReasons[claim.discountFactor] ??
      "This facility's verification status determines how much of its capacity can back a claim.",
    referenceFactorValue: claim.referenceFactorValue,
    referenceFactorUnit: "kgCO2e/kWh",
    referenceTableVersion: claim.referenceFactorId,
  },
  status: claim.status,
  priority: claim.priority,
  requiresDualApproval: claim.requiresDualApproval,
  evidence: evidenceFrom(claim),
  decisions: [],
  issuedAmount: claim.issuedAmount ? creditAmount(claim.issuedAmount) : null,
  exclusivityAttestedAt: claim.exclusivityAttestedAt,
  submittedAt: claim.createdAt,
});
