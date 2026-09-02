import "server-only";
import { gatewayGet, gatewayPost } from "@/lib/api/gateway";
import type { FacilityVerificationStatus, TrustTier } from "@/lib/status";
import type { FacilityType, GridRegion } from "@/lib/types";

type ApiFacility = {
  id: string;
  name: string;
  address: string;
  country_code: string;
  grid_region: GridRegion;
  type: FacilityType;
  facility_reference: string | null;
  verification_status: FacilityVerificationStatus;
  ceiling_discount_factor: string;
  trust_tier: TrustTier;
  declared_capacity: string;
  declared_energy_kwh: string;
  attested_capacity: string | null;
  attested_energy_kwh: string | null;
  created_at: string;
};

export type FacilityRecord = {
  id: string;
  name: string;
  address: string;
  countryCode: string;
  gridRegion: GridRegion;
  type: FacilityType;
  facilityReference: string | null;
  verificationStatus: FacilityVerificationStatus;
  ceilingDiscountFactor: string;
  trustTier: TrustTier;
  declaredAnnualProductionCapacity: string;
  declaredAnnualEnergyConsumptionKwh: string;
  attestedAnnualProductionCapacity: string | null;
  attestedAnnualEnergyConsumptionKwh: string | null;
  createdAt: string;
};

export type FacilityDraft = {
  name: string;
  address: string;
  countryCode: string;
  gridRegion: GridRegion;
  type: FacilityType;
  facilityReference: string;
  declaredAnnualProductionCapacity: string;
  declaredAnnualEnergyConsumptionKwh: string;
};

const toFacility = (facility: ApiFacility): FacilityRecord => ({
  id: facility.id,
  name: facility.name,
  address: facility.address,
  countryCode: facility.country_code,
  gridRegion: facility.grid_region,
  type: facility.type,
  facilityReference: facility.facility_reference,
  verificationStatus: facility.verification_status,
  ceilingDiscountFactor: facility.ceiling_discount_factor,
  trustTier: facility.trust_tier,
  declaredAnnualProductionCapacity: facility.declared_capacity,
  declaredAnnualEnergyConsumptionKwh: facility.declared_energy_kwh,
  attestedAnnualProductionCapacity: facility.attested_capacity,
  attestedAnnualEnergyConsumptionKwh: facility.attested_energy_kwh,
  createdAt: facility.created_at,
});

export const fetchFacilities = async (
  token: string,
): Promise<FacilityRecord[]> => {
  const listed = await gatewayGet<{ facilities: ApiFacility[] }>(
    "/v1/facilities",
    token,
  );
  return listed.facilities.map(toFacility);
};

export const fetchFacility = async (
  token: string,
  facilityId: string,
): Promise<FacilityRecord> =>
  toFacility(
    await gatewayGet<ApiFacility>(`/v1/facilities/${facilityId}`, token),
  );

export const createFacility = async (
  token: string,
  draft: FacilityDraft,
  idempotencyKey: string,
): Promise<FacilityRecord> => {
  const created = await gatewayPost<ApiFacility>(
    "/v1/facilities",
    token,
    {
      name: draft.name,
      address: draft.address,
      country_code: draft.countryCode,
      grid_region: draft.gridRegion,
      type: draft.type,
      facility_reference: draft.facilityReference,
      declared_capacity: draft.declaredAnnualProductionCapacity,
      declared_energy_kwh: draft.declaredAnnualEnergyConsumptionKwh,
    },
    idempotencyKey,
  );

  return toFacility(created);
};
