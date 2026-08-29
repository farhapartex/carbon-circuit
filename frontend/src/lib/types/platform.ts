import type { FraudSeverity } from "@/lib/status";
import type { Id, IsoTimestamp } from "@/lib/types/common";

export type FraudFlagState = "open" | "reviewed" | "escalated" | "resolved";

export type FraudSubjectKind =
  "batch" | "checkpoint" | "claim" | "evidence" | "trade";

export type FraudFlag = {
  id: Id;
  ruleId: string;
  ruleLabel: string;
  severity: FraudSeverity;
  state: FraudFlagState;
  subjectKind: FraudSubjectKind;
  subjectId: Id;
  computedEvidence: string;
  raisedAt: IsoTimestamp;
};

export type NotificationCategory =
  | "verification"
  | "provenance"
  | "claim"
  | "credit"
  | "marketplace"
  | "billing"
  | "security"
  | "fraud";

export type Notification = {
  id: Id;
  category: NotificationCategory;
  title: string;
  body: string;
  href: string | null;
  read: boolean;
  occurredAt: IsoTimestamp;
  collapsedCount: number;
};

export type DataExportRequest = {
  id: Id;
  state: "processing" | "ready" | "expired" | "failed";
  requestedAt: IsoTimestamp;
  downloadUrl: string | null;
  expiresAt: IsoTimestamp | null;
};

export const NOTIFICATION_DIGEST_THRESHOLD = 10;
export const DATA_EXPORT_COOLDOWN_HOURS = 24;
export const DATA_EXPORT_LINK_VALID_DAYS = 7;
