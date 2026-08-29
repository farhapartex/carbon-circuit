export type Id = string;
export type IsoTimestamp = string;
export type EthereumAddress = `0x${string}`;
export type TransactionHash = `0x${string}`;
export type PublicBatchReference = string;
export type CountryCode = string;

export type PageMeta = {
  page: number;
  perPage: number;
  totalItems: number;
  totalPages: number;
};

export type CursorMeta = {
  nextCursor: string | null;
  hasMore: boolean;
};

export type Paginated<T> = {
  items: T[];
  meta: PageMeta;
};

export type CursorPaginated<T> = {
  items: T[];
  meta: CursorMeta;
};

export type ProductCategory =
  "electronics" | "agriculture" | "pharma" | "textiles";

export type ActivityType =
  "renewable_energy" | "reduced_emission_logistics" | "responsible_sourcing";

export type GridRegion =
  | "US-CAISO"
  | "US-ERCOT"
  | "US-PJM"
  | "US-MISO"
  | "EU-DE"
  | "EU-FR"
  | "EU-PL"
  | "UK"
  | "CN-East"
  | "CN-South"
  | "IN-North"
  | "JP"
  | "KR"
  | "TW"
  | "VN"
  | "MY"
  | "SG"
  | "TH";

export type ShippingMethod =
  | "air_freight_short_haul"
  | "air_freight_long_haul"
  | "sea_freight_container"
  | "sea_freight_bulk"
  | "rail_electric"
  | "rail_diesel"
  | "road_hgv"
  | "road_lgv"
  | "inland_waterway";

export type RecycledMaterial =
  | "aluminium"
  | "copper"
  | "steel"
  | "tin"
  | "gold"
  | "tantalum"
  | "plastics_abs"
  | "plastics_pet"
  | "rare_earth_magnets"
  | "lithium_black_mass";

export type GeoPoint = {
  latitude: number;
  longitude: number;
};

export type Location = {
  label: string;
  countryCode: CountryCode;
  coordinates: GeoPoint | null;
};
