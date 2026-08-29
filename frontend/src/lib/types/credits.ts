import type { CreditAmount } from "@/lib/decimal";
import type {
  ActivityType,
  CountryCode,
  Id,
  IsoTimestamp,
  TransactionHash,
} from "@/lib/types/common";

export type CreditClass = {
  tokenId: string;
  facilityId: Id;
  facilityName: string;
  facilityCountry: CountryCode;
  vintageYear: number;
  activityType: ActivityType;
};

export type CreditIssuance = {
  id: Id;
  claimId: Id;
  creditClass: CreditClass;
  amount: CreditAmount;
  transactionHash: TransactionHash | null;
  issuedAt: IsoTimestamp;
};

export type CreditClassBalance = {
  creditClass: CreditClass;
  available: CreditAmount;
  escrowed: CreditAmount;
  retired: CreditAmount;
};

export type CreditPortfolio = {
  balances: CreditClassBalance[];
  totalAvailable: CreditAmount;
  totalEscrowed: CreditAmount;
  totalRetired: CreditAmount;
};
