import type { CreditAmount } from "@/lib/decimal";
import type {
  FacilityVerificationStatus,
  TrustTier,
  VerificationStatus,
} from "@/lib/status";
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

export type UserProfile = {
  id: Id;
  name: string;
  email: string;
  emailVerified: boolean;
  role: OrganizationRole | null;
  platformRole: PlatformRole | null;
  personalWalletAddress: EthereumAddress | null;
  createdAt: IsoTimestamp;
};

export type MfaMethodKind = "authenticator_app" | "sms" | "recovery_codes";

export type MfaMethod = {
  kind: MfaMethodKind;
  label: string;
  detail: string | null;
  enrolledAt: IsoTimestamp | null;
  isDefault: boolean;
};

export type MfaSettings = {
  enabled: boolean;
  requiredByRole: boolean;
  methods: MfaMethod[];
  recoveryCodesRemaining: number;
  lastVerifiedAt: IsoTimestamp | null;
};

export type InvitationState = "pending" | "accepted" | "revoked" | "expired";

export type OrganizationInvitation = {
  id: Id;
  email: string;
  role: OrganizationRole;
  state: InvitationState;
  invitedByName: string;
  invitedAt: IsoTimestamp;
  expiresAt: IsoTimestamp;
};

export type RegistryEntityStatus = "active" | "dissolved";

export type BusinessRegistryRecord = {
  countryCode: CountryCode;
  registrationNumber: string;
  legalName: string;
  registeredAddress: string;
  incorporationDate: IsoTimestamp;
  entityStatus: RegistryEntityStatus;
  industryCodes: string[];
  sanctioned: boolean;
};

export type RegistryRejectionReason =
  "entity_dissolved" | "sanctions_flag" | "name_mismatch";

export type RegistryVerificationOutcome = {
  status: VerificationStatus;
  matchedRecord: BusinessRegistryRecord | null;
  nameSimilarity: number | null;
  rejectionReason: RegistryRejectionReason | null;
};

export type OnboardingStep =
  "organization" | "verification" | "plan" | "wallet" | "complete";
