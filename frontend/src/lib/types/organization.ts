import type { CreditAmount } from "@/lib/decimal";
import type { TrustTier, VerificationStatus } from "@/lib/status";
import type {
  CountryCode,
  EthereumAddress,
  GridRegion,
  Id,
  IsoTimestamp,
  ProductCategory,
} from "@/lib/types/common";

export type OrganizationType =
  "manufacturer" | "assembler" | "logistics" | "credit_buyer";

export type OrganizationState =
  "active" | "restricted" | "read_only" | "suspended";

export type OrganizationRole = "owner" | "admin" | "member";

export type PlatformRole = "verifier" | "platform_admin";

export type Organization = {
  id: Id;
  name: string;
  type: OrganizationType;
  countryOfIncorporation: CountryCode;
  businessRegistrationNumber: string;
  verificationStatus: VerificationStatus;
  state: OrganizationState;
  productCategories: ProductCategory[];
  treasuryAddress: EthereumAddress | null;
  createdAt: IsoTimestamp;
};

export type OrganizationUser = {
  id: Id;
  name: string;
  email: string;
  role: OrganizationRole;
  mfaEnabled: boolean;
  lastActiveAt: IsoTimestamp | null;
  invitedAt: IsoTimestamp;
};

export type FacilityType =
  | "raw_material_processing"
  | "component_fabrication"
  | "assembly"
  | "distribution";

export type FacilityVerificationStatus =
  "facility_matched" | "organization_matched" | "self_declared";

export type Facility = {
  id: Id;
  organizationId: Id;
  name: string;
  address: string;
  countryCode: CountryCode;
  gridRegion: GridRegion;
  type: FacilityType;
  verificationStatus: FacilityVerificationStatus;
  ceilingDiscountFactor: "1.00" | "0.75" | "0.50";
  trustTier: TrustTier;
  declaredAnnualProductionCapacity: string;
  declaredAnnualEnergyConsumptionKwh: string;
  attestedAnnualProductionCapacity: string | null;
  attestedAnnualEnergyConsumptionKwh: string | null;
  batchCount: number;
  approvedClaimCount: number;
  createdAt: IsoTimestamp;
};

export type TrustTierCriteria = {
  decidedClaimCount: number;
  approvalRate: number;
  organizationVerified: boolean;
  facilityRegistryMatched: boolean;
  distinctActivityTypes: number;
  escalatedFraudFlagsInWindow: number;
};

export type TreasuryAddressChangeState = "pending" | "completed" | "cancelled";

export type TreasuryAddressChange = {
  id: Id;
  organizationId: Id;
  currentAddress: EthereumAddress | null;
  requestedAddress: EthereumAddress;
  state: TreasuryAddressChangeState;
  initiatedBy: string;
  initiatedAt: IsoTimestamp;
  effectiveAt: IsoTimestamp;
  resolvedAt: IsoTimestamp | null;
  escrowedCreditsBlocking: CreditAmount | null;
};

export type ApiKey = {
  id: Id;
  name: string;
  prefix: string;
  createdAt: IsoTimestamp;
  lastUsedAt: IsoTimestamp | null;
  revokedAt: IsoTimestamp | null;
};

export type ActiveSession = {
  id: Id;
  userAgent: string;
  ipAddress: string;
  startedAt: IsoTimestamp;
  lastSeenAt: IsoTimestamp;
  current: boolean;
};
