import type { FraudFlag, Notification } from "@/lib/types";

export const notifications: Notification[] = [
  {
    id: "ntf_credits_issued",
    category: "credit",
    title: "6,016.92 tCO2e issued to your Treasury Address",
    body: "Credits from claim clm_renewable_2026_q2 have been minted for Hsinchu Fab TW-01, vintage 2026, renewable energy.",
    href: "/credits",
    read: false,
    occurredAt: "2026-08-12T13:26:00Z",
    collapsedCount: 1,
  },
  {
    id: "ntf_treasury_change",
    category: "security",
    title: "Treasury Address change requested",
    body: "Wei-Chen Lin requested a change to 0x7b3c…7b9c. This takes effect on 31 August 2026 and any Owner can cancel it until then.",
    href: "/organization/wallet",
    read: false,
    occurredAt: "2026-08-28T10:00:00Z",
    collapsedCount: 1,
  },
  {
    id: "ntf_listing_sold",
    category: "marketplace",
    title: "Your listing sold",
    body: "787.25 tCO2e of Hsinchu Fab TW-01 vintage 2025 credits sold to Halden Insurance Group.",
    href: "/marketplace/my-listings",
    read: true,
    occurredAt: "2026-07-30T14:22:00Z",
    collapsedCount: 1,
  },
  {
    id: "ntf_checkpoint_digest",
    category: "provenance",
    title: "142 checkpoints logged",
    body: "Bulk ingest from ERP checkpoint ingest completed across 38 batches.",
    href: "/batches",
    read: true,
    occurredAt: "2026-08-29T06:15:00Z",
    collapsedCount: 142,
  },
  {
    id: "ntf_quota_warning",
    category: "billing",
    title: "Checkpoint usage at 89% of your plan limit",
    body: "17,840 of 20,000 checkpoints used this billing period, which ends on 11 September 2026.",
    href: "/billing",
    read: true,
    occurredAt: "2026-08-27T02:00:00Z",
    collapsedCount: 1,
  },
];

export const fraudFlags: FraudFlag[] = [
  {
    id: "frd_impossible_travel",
    ruleId: "impossible_travel",
    ruleLabel: "Impossible travel",
    severity: "medium",
    state: "reviewed",
    subjectKind: "checkpoint",
    subjectId: "chk_p3",
    computedEvidence:
      "Implied speed of 1,180 km/h between consecutive checkpoints exceeds the 900 km/h air freight ceiling.",
    raisedAt: "2026-08-28T18:44:00Z",
  },
  {
    id: "frd_backdated",
    ruleId: "backdated_checkpoint",
    ruleLabel: "Backdated checkpoint",
    severity: "low",
    state: "open",
    subjectKind: "checkpoint",
    subjectId: "chk_p2",
    computedEvidence:
      "Event timestamp is 44 days before submission time, beyond the 30 day threshold.",
    raisedAt: "2026-08-20T14:10:00Z",
  },
];
