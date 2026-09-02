import type { ShippingMethod } from "@/lib/types/common";
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
