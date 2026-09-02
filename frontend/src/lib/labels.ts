import type { FacilityVerificationStatus } from "@/lib/status";
import type {
  ActivityType,
  GridRegion,
  ProductCategory,
  RecycledMaterial,
  ShippingMethod,
} from "@/lib/types/common";
import type { FacilityType } from "@/lib/types/organization";
import type { CheckpointType } from "@/lib/types/provenance";

export const checkpointTypeLabels: Record<CheckpointType, string> = {
  production_complete: "Production complete",
  departed_origin: "Departed origin",
  customs_export: "Cleared export customs",
  customs_import: "Cleared import customs",
  arrived_destination: "Arrived at destination",
};

export const shippingMethodLabels: Record<ShippingMethod, string> = {
  air_freight_short_haul: "Air freight, short-haul",
  air_freight_long_haul: "Air freight, long-haul",
  sea_freight_container: "Sea freight, container",
  sea_freight_bulk: "Sea freight, bulk",
  rail_electric: "Rail, electric",
  rail_diesel: "Rail, diesel",
  road_hgv: "Road, heavy goods vehicle",
  road_lgv: "Road, light goods vehicle",
  inland_waterway: "Inland waterway",
};

export const movesGoods = (type: CheckpointType) =>
  type !== "production_complete";

export const facilityTypeLabels: Record<FacilityType, string> = {
  raw_material_processing: "Raw material processing",
  component_fabrication: "Component fabrication",
  assembly: "Assembly",
  distribution: "Distribution",
};

export const discountFactorRationale: Record<
  FacilityVerificationStatus,
  string
> = {
  facility_matched:
    "This facility matched the registry dataset, so its attested capacity is used and no discount applies.",
  organization_matched:
    "Your organization matched the business registry but this facility did not match individually, so its declared scale carries a 25% discount.",
  self_declared:
    "Nothing about this facility's declared scale is independently corroborated, so its ceiling is limited to half of what the declared capacity would support.",
};

export const activityTypeLabels: Record<ActivityType, string> = {
  renewable_energy: "Renewable energy",
  reduced_emission_logistics: "Reduced-emission logistics",
  responsible_sourcing: "Responsible sourcing",
};

export const gridRegionLabels: Record<GridRegion, string> = {
  "US-CAISO": "US-CAISO (California)",
  "US-ERCOT": "US-ERCOT (Texas)",
  "US-PJM": "US-PJM (Mid-Atlantic)",
  "US-MISO": "US-MISO (Midwest)",
  "EU-DE": "EU-DE (Germany)",
  "EU-FR": "EU-FR (France)",
  "EU-PL": "EU-PL (Poland)",
  UK: "UK (United Kingdom)",
  "CN-East": "CN-East (Eastern China)",
  "CN-South": "CN-South (Southern China)",
  "IN-North": "IN-North (Northern India)",
  JP: "JP (Japan)",
  KR: "KR (South Korea)",
  TW: "TW (Taiwan)",
  VN: "VN (Vietnam)",
  MY: "MY (Malaysia)",
  SG: "SG (Singapore)",
  TH: "TH (Thailand)",
};

export const productCategoryLabels: Record<ProductCategory, string> = {
  electronics: "Electronics",
  agriculture: "Agriculture",
  pharma: "Pharma",
  textiles: "Textiles",
};

export const recycledMaterialLabels: Record<RecycledMaterial, string> = {
  aluminium: "Aluminium",
  copper: "Copper",
  steel: "Steel",
  tin: "Tin",
  gold: "Gold",
  tantalum: "Tantalum",
  plastics_abs: "ABS plastics",
  plastics_pet: "PET plastics",
  rare_earth_magnets: "Rare earth magnets",
  lithium_black_mass: "Lithium black mass",
};
