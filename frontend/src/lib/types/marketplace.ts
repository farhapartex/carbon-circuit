import type { CreditAmount, UsdcAmount } from "@/lib/decimal";
import type { ListingStatus, TrustTier } from "@/lib/status";
import type { Id, IsoTimestamp, TransactionHash } from "@/lib/types/common";
import type { CreditClass } from "@/lib/types/credits";

export type MarketplaceListing = {
  id: Id;
  sellerOrganizationId: Id;
  sellerOrganizationName: string;
  sellerTrustTier: TrustTier;
  creditClass: CreditClass;
  quantityRemaining: CreditAmount;
  quantityOriginal: CreditAmount;
  pricePerTonne: UsdcAmount;
  minimumPurchaseQuantity: CreditAmount;
  status: ListingStatus;
  expiresAt: IsoTimestamp;
  createdAt: IsoTimestamp;
};

export type PurchaseQuote = {
  quantity: CreditAmount;
  cost: UsdcAmount;
  fee: UsdcAmount;
  sellerProceeds: UsdcAmount;
};

export type ListingProceedsPreview = {
  gross: UsdcAmount;
  feeBasisPoints: number;
  fee: UsdcAmount;
  net: UsdcAmount;
};

export type Trade = {
  id: Id;
  listingId: Id;
  creditClass: CreditClass;
  quantity: CreditAmount;
  cost: UsdcAmount;
  fee: UsdcAmount;
  buyerOrganizationName: string;
  sellerOrganizationName: string;
  transactionHash: TransactionHash | null;
  settledAt: IsoTimestamp;
};

export type Retirement = {
  id: Id;
  retiringOrganizationName: string;
  creditClass: CreditClass;
  quantity: CreditAmount;
  purpose: string;
  transactionHash: TransactionHash | null;
  retiredAt: IsoTimestamp;
};

export const MINIMUM_LISTING_QUANTITY = "1.000000";
export const MINIMUM_PURCHASE_QUANTITY_FLOOR = "0.100000";
export const MINIMUM_TRANSACTION_NOTIONAL_USDC = "1.00";
export const MINIMUM_PRICE_PER_TONNE_USDC = "0.50";
export const MAXIMUM_PRICE_PER_TONNE_USDC = "5000.00";
export const LISTING_MAXIMUM_DURATION_DAYS = 90;
export const RETIREMENT_PURPOSE_MINIMUM_LENGTH = 20;
