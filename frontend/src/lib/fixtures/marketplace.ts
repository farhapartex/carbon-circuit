import { creditAmount, usdcAmount } from "@/lib/decimal";
import {
  facilityMatched,
  organizationMatched,
  verifiedManufacturer,
} from "@/lib/fixtures/organizations";
import type {
  CreditClass,
  CreditPortfolio,
  MarketplaceListing,
  Retirement,
  Trade,
} from "@/lib/types";

export const renewableTw2026: CreditClass = {
  tokenId: "0x5c8e1f3a7b9d2e4f6a8c0b2d4e6f8a0c2e4b6d8f0a2c4e6b8d0f2a4c6e8b0d2f",
  facilityId: facilityMatched.id,
  facilityName: facilityMatched.name,
  facilityCountry: "TW",
  vintageYear: 2026,
  activityType: "renewable_energy",
};

export const renewableTw2025: CreditClass = {
  tokenId: "0x1a3c5e7b9d1f3a5c7e9b1d3f5a7c9e1b3d5f7a9c1e3b5d7f9a1c3e5b7d9f1a3c",
  facilityId: facilityMatched.id,
  facilityName: facilityMatched.name,
  facilityCountry: "TW",
  vintageYear: 2025,
  activityType: "renewable_energy",
};

export const logisticsTw2026: CreditClass = {
  tokenId: "0x9e1b3d5f7a9c1e3b5d7f9a1c3e5b7d9f1a3c5e7b9d1f3a5c7e9b1d3f5a7c9e1b",
  facilityId: organizationMatched.id,
  facilityName: organizationMatched.name,
  facilityCountry: "TW",
  vintageYear: 2026,
  activityType: "reduced_emission_logistics",
};

export const creditPortfolio: CreditPortfolio = {
  balances: [
    {
      creditClass: renewableTw2026,
      available: creditAmount("3016.920000"),
      escrowed: creditAmount("3000.000000"),
      retired: creditAmount("0"),
    },
    {
      creditClass: renewableTw2025,
      available: creditAmount("1840.500000"),
      escrowed: creditAmount("0"),
      retired: creditAmount("250.000000"),
    },
    {
      creditClass: logisticsTw2026,
      available: creditAmount("64.250000"),
      escrowed: creditAmount("0"),
      retired: creditAmount("0"),
    },
  ],
  totalAvailable: creditAmount("4921.670000"),
  totalEscrowed: creditAmount("3000.000000"),
  totalRetired: creditAmount("250.000000"),
};

export const listings: MarketplaceListing[] = [
  {
    id: "lst_renewable_tw_2026",
    sellerOrganizationId: verifiedManufacturer.id,
    sellerOrganizationName: verifiedManufacturer.name,
    sellerTrustTier: "trusted",
    creditClass: renewableTw2026,
    quantityRemaining: creditAmount("3000.000000"),
    quantityOriginal: creditAmount("3000.000000"),
    pricePerTonne: usdcAmount("18.50"),
    minimumPurchaseQuantity: creditAmount("10.000000"),
    status: "active",
    expiresAt: "2026-11-12T00:00:00Z",
    createdAt: "2026-08-14T10:00:00Z",
  },
  {
    id: "lst_renewable_tw_2025",
    sellerOrganizationId: verifiedManufacturer.id,
    sellerOrganizationName: verifiedManufacturer.name,
    sellerTrustTier: "trusted",
    creditClass: renewableTw2025,
    quantityRemaining: creditAmount("412.750000"),
    quantityOriginal: creditAmount("1200.000000"),
    pricePerTonne: usdcAmount("14.25"),
    minimumPurchaseQuantity: creditAmount("5.000000"),
    status: "partially_filled",
    expiresAt: "2026-09-03T00:00:00Z",
    createdAt: "2026-06-05T09:30:00Z",
  },
  {
    id: "lst_logistics_tw_2026",
    sellerOrganizationId: verifiedManufacturer.id,
    sellerOrganizationName: verifiedManufacturer.name,
    sellerTrustTier: "verified",
    creditClass: logisticsTw2026,
    quantityRemaining: creditAmount("40.000000"),
    quantityOriginal: creditAmount("40.000000"),
    pricePerTonne: usdcAmount("31.00"),
    minimumPurchaseQuantity: creditAmount("1.000000"),
    status: "active",
    expiresAt: "2026-10-20T00:00:00Z",
    createdAt: "2026-07-22T16:12:00Z",
  },
];

export const trades: Trade[] = [
  {
    id: "trd_0001",
    listingId: "lst_renewable_tw_2025",
    creditClass: renewableTw2025,
    quantity: creditAmount("787.250000"),
    cost: usdcAmount("11218.32"),
    fee: usdcAmount("280.45"),
    buyerOrganizationName: "Halden Insurance Group",
    sellerOrganizationName: verifiedManufacturer.name,
    transactionHash:
      "0x3f5a7c9e1b3d5f7a9c1e3b5d7f9a1c3e5b7d9f1a3c5e7b9d1f3a5c7e9b1d3f5a",
    settledAt: "2026-07-30T14:22:00Z",
  },
];

export const retirements: Retirement[] = [
  {
    id: "ret_0001",
    retiringOrganizationName: "Halden Insurance Group",
    creditClass: renewableTw2025,
    quantity: creditAmount("250.000000"),
    purpose: "2026 annual ESG report offset, Scope 2 residual emissions",
    transactionHash:
      "0x7c9e1b3d5f7a9c1e3b5d7f9a1c3e5b7d9f1a3c5e7b9d1f3a5c7e9b1d3f5a7c9e",
    retiredAt: "2026-08-02T09:45:00Z",
  },
  {
    id: "ret_0002",
    retiringOrganizationName: "Deleted Organization #4417",
    creditClass: renewableTw2025,
    quantity: creditAmount("120.500000"),
    purpose: "Voluntary offset against 2025 corporate travel",
    transactionHash:
      "0x1b3d5f7a9c1e3b5d7f9a1c3e5b7d9f1a3c5e7b9d1f3a5c7e9b1d3f5a7c9e1b3d",
    retiredAt: "2026-05-19T11:10:00Z",
  },
  {
    id: "ret_0003",
    retiringOrganizationName: "Wrenfield Logistics Holdings",
    creditClass: logisticsTw2026,
    quantity: creditAmount("18.000000"),
    purpose: "Offsetting Q2 2026 inbound freight emissions",
    transactionHash:
      "0x5f7a9c1e3b5d7f9a1c3e5b7d9f1a3c5e7b9d1f3a5c7e9b1d3f5a7c9e1b3d5f7a",
    retiredAt: "2026-08-21T15:38:00Z",
  },
];
