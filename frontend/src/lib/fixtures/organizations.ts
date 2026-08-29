import type {
  ActiveSession,
  ApiKey,
  Facility,
  MfaSettings,
  Organization,
  OrganizationInvitation,
  OrganizationUser,
  TreasuryAddressChange,
  UserProfile,
} from "@/lib/types";

export const verifiedManufacturer: Organization = {
  id: "org_tw_semiconductor",
  name: "Formosa Precision Semiconductor Co., Ltd.",
  type: "manufacturer",
  countryOfIncorporation: "TW",
  businessRegistrationNumber: "TW-28419377",
  verificationStatus: "verified",
  state: "active",
  productCategories: ["electronics"],
  treasuryAddress: "0x4a7f2b1c9e3d5a8f6b0c2d4e6f8a0b2c4d6e8f01",
  createdAt: "2025-11-04T09:12:00Z",
};

export const unverifiedAssembler: Organization = {
  id: "org_unverified_assembler",
  name: "Northbridge Assembly Works",
  type: "assembler",
  countryOfIncorporation: "MY",
  businessRegistrationNumber: "MY-9920551",
  verificationStatus: "unverified",
  state: "active",
  productCategories: ["electronics"],
  treasuryAddress: null,
  createdAt: "2026-06-18T14:41:00Z",
};

export const restrictedLogisticsPartner: Organization = {
  id: "org_restricted_logistics",
  name: "Meridian Freight Solutions",
  type: "logistics",
  countryOfIncorporation: "SG",
  businessRegistrationNumber: "SG-201744820K",
  verificationStatus: "verified",
  state: "restricted",
  productCategories: ["electronics"],
  treasuryAddress: "0x9c1d3e5f7a9b1c3d5e7f9a1b3c5d7e9f1a3b5c7d",
  createdAt: "2025-08-22T07:30:00Z",
};

export const creditBuyer: Organization = {
  id: "org_credit_buyer",
  name: "Halden Insurance Group",
  type: "credit_buyer",
  countryOfIncorporation: "UK",
  businessRegistrationNumber: "UK-08812340",
  verificationStatus: "verified",
  state: "active",
  productCategories: [],
  treasuryAddress: "0x2f4a6b8c0d2e4f6a8b0c2d4e6f8a0b2c4d6e8f0a",
  createdAt: "2026-02-09T11:05:00Z",
};

export const facilityMatched: Facility = {
  id: "fac_tw_01",
  organizationId: verifiedManufacturer.id,
  name: "Hsinchu Fab TW-01",
  address: "No. 8 Li-Hsin Road, Hsinchu Science Park",
  countryCode: "TW",
  gridRegion: "TW",
  type: "component_fabrication",
  verificationStatus: "facility_matched",
  ceilingDiscountFactor: "1.00",
  trustTier: "trusted",
  declaredAnnualProductionCapacity: "18000000",
  declaredAnnualEnergyConsumptionKwh: "31000000",
  attestedAnnualProductionCapacity: "18000000",
  attestedAnnualEnergyConsumptionKwh: "31000000",
  batchCount: 42,
  approvedClaimCount: 11,
  createdAt: "2025-11-04T09:20:00Z",
};

export const organizationMatched: Facility = {
  id: "fac_tw_02",
  organizationId: verifiedManufacturer.id,
  name: "Tainan Packaging TW-02",
  address: "No. 21 Nanke 3rd Road, Tainan Science Park",
  countryCode: "TW",
  gridRegion: "TW",
  type: "assembly",
  verificationStatus: "organization_matched",
  ceilingDiscountFactor: "0.75",
  trustTier: "verified",
  declaredAnnualProductionCapacity: "9500000",
  declaredAnnualEnergyConsumptionKwh: "12400000",
  attestedAnnualProductionCapacity: null,
  attestedAnnualEnergyConsumptionKwh: null,
  batchCount: 17,
  approvedClaimCount: 4,
  createdAt: "2026-01-15T10:02:00Z",
};

export const selfDeclared: Facility = {
  id: "fac_my_01",
  organizationId: unverifiedAssembler.id,
  name: "Penang Line MY-01",
  address: "Plot 14, Bayan Lepas Free Industrial Zone",
  countryCode: "MY",
  gridRegion: "MY",
  type: "assembly",
  verificationStatus: "self_declared",
  ceilingDiscountFactor: "0.50",
  trustTier: "new",
  declaredAnnualProductionCapacity: "4200000",
  declaredAnnualEnergyConsumptionKwh: "6800000",
  attestedAnnualProductionCapacity: null,
  attestedAnnualEnergyConsumptionKwh: null,
  batchCount: 3,
  approvedClaimCount: 0,
  createdAt: "2026-06-18T15:00:00Z",
};

export const facilities: Facility[] = [
  facilityMatched,
  organizationMatched,
  selfDeclared,
];

export const organizationUsers: OrganizationUser[] = [
  {
    id: "usr_owner",
    name: "Wei-Chen Lin",
    email: "wc.lin@formosaprecision.example",
    role: "owner",
    mfaEnabled: true,
    lastActiveAt: "2026-08-29T08:14:00Z",
    invitedAt: "2025-11-04T09:12:00Z",
  },
  {
    id: "usr_admin",
    name: "Priya Raghavan",
    email: "p.raghavan@formosaprecision.example",
    role: "admin",
    mfaEnabled: true,
    lastActiveAt: "2026-08-28T16:52:00Z",
    invitedAt: "2025-12-01T13:20:00Z",
  },
  {
    id: "usr_member",
    name: "Tomas Eriksson",
    email: "t.eriksson@formosaprecision.example",
    role: "member",
    mfaEnabled: false,
    lastActiveAt: null,
    invitedAt: "2026-08-20T09:00:00Z",
  },
];

export const apiKeys: ApiKey[] = [
  {
    id: "key_erp",
    name: "ERP checkpoint ingest",
    prefix: "cc_live_8f3a2b1c",
    createdAt: "2026-03-11T09:00:00Z",
    lastUsedAt: "2026-08-29T07:58:00Z",
    revokedAt: null,
  },
  {
    id: "key_retired",
    name: "Legacy WMS bridge",
    prefix: "cc_live_1d5e7f90",
    createdAt: "2025-12-02T11:30:00Z",
    lastUsedAt: "2026-05-14T22:10:00Z",
    revokedAt: "2026-05-15T09:00:00Z",
  },
];

export const pendingTreasuryChange: TreasuryAddressChange = {
  id: "tac_pending",
  organizationId: verifiedManufacturer.id,
  currentAddress: verifiedManufacturer.treasuryAddress,
  requestedAddress: "0x7b3c5d7e9f1a3b5c7d9e1f3a5b7c9d1e3f5a7b9c",
  state: "pending",
  initiatedBy: "Wei-Chen Lin",
  initiatedAt: "2026-08-28T10:00:00Z",
  effectiveAt: "2026-08-31T10:00:00Z",
  resolvedAt: null,
  escrowedCreditsBlocking: null,
};

export const activeSessions: ActiveSession[] = [
  {
    id: "sess_current",
    userAgent: "Chrome 141 on macOS",
    ipAddress: "203.0.113.42",
    startedAt: "2026-08-29T07:45:00Z",
    lastSeenAt: "2026-08-29T08:14:00Z",
    current: true,
  },
  {
    id: "sess_other",
    userAgent: "Safari 19 on iOS",
    ipAddress: "198.51.100.7",
    startedAt: "2026-08-27T19:02:00Z",
    lastSeenAt: "2026-08-28T21:33:00Z",
    current: false,
  },
];

export const signedInUserProfile: UserProfile = {
  id: "usr_owner",
  name: "Wei-Chen Lin",
  email: "wc.lin@formosaprecision.example",
  emailVerified: true,
  role: "owner",
  platformRole: null,
  personalWalletAddress: "0xd4e6f8a0b2c4d6e8f0a2b4c6d8e0f2a4b6c8d0e2",
  createdAt: "2025-11-04T09:12:00Z",
};

export const mfaSettings: MfaSettings = {
  enabled: true,
  requiredByRole: true,
  methods: [
    {
      kind: "authenticator_app",
      label: "Authenticator app",
      detail: "Added from iPhone 17 Pro",
      enrolledAt: "2025-11-04T09:31:00Z",
      isDefault: true,
    },
    {
      kind: "sms",
      label: "SMS backup",
      detail: "Ending 4417",
      enrolledAt: "2025-11-04T09:34:00Z",
      isDefault: false,
    },
    {
      kind: "recovery_codes",
      label: "Recovery codes",
      detail: "8 of 10 remaining",
      enrolledAt: "2025-11-04T09:36:00Z",
      isDefault: false,
    },
  ],
  recoveryCodesRemaining: 8,
  lastVerifiedAt: "2026-08-28T10:00:00Z",
};

export const invitations: OrganizationInvitation[] = [
  {
    id: "inv_pending_verifier_liaison",
    email: "s.okafor@formosaprecision.example",
    role: "admin",
    state: "pending",
    invitedByName: "Wei-Chen Lin",
    invitedAt: "2026-08-26T11:20:00Z",
    expiresAt: "2026-09-02T11:20:00Z",
  },
  {
    id: "inv_pending_analyst",
    email: "m.tanaka@formosaprecision.example",
    role: "member",
    state: "pending",
    invitedByName: "Priya Raghavan",
    invitedAt: "2026-08-28T09:05:00Z",
    expiresAt: "2026-09-04T09:05:00Z",
  },
  {
    id: "inv_expired",
    email: "old.contractor@formosaprecision.example",
    role: "member",
    state: "expired",
    invitedByName: "Wei-Chen Lin",
    invitedAt: "2026-07-01T08:00:00Z",
    expiresAt: "2026-07-08T08:00:00Z",
  },
];
