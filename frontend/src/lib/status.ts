export type StatusVariant =
  "neutral" | "primary" | "success" | "warning" | "danger" | "info";

export type StatusPresentation = {
  label: string;
  variant: StatusVariant;
};

export type ClaimStatus =
  | "draft"
  | "submitted"
  | "under_ai_review"
  | "under_human_review"
  | "approved"
  | "rejected"
  | "more_information_requested";

export const claimStatusPresentation: Record<ClaimStatus, StatusPresentation> =
  {
    draft: { label: "Draft", variant: "neutral" },
    submitted: { label: "Submitted", variant: "warning" },
    under_ai_review: { label: "Under AI Review", variant: "warning" },
    under_human_review: { label: "Under Human Review", variant: "warning" },
    approved: { label: "Approved", variant: "success" },
    rejected: { label: "Rejected", variant: "danger" },
    more_information_requested: {
      label: "More Information Requested",
      variant: "info",
    },
  };

export type ListingStatus =
  "active" | "partially_filled" | "filled" | "cancelled" | "expired";

export const listingStatusPresentation: Record<
  ListingStatus,
  StatusPresentation
> = {
  active: { label: "Active", variant: "success" },
  partially_filled: { label: "Partially filled", variant: "info" },
  filled: { label: "Filled", variant: "neutral" },
  cancelled: { label: "Cancelled", variant: "neutral" },
  expired: { label: "Expired", variant: "neutral" },
};

export type TransactionStatus =
  "awaiting_signature" | "pending" | "confirmed" | "failed";

export const transactionStatusPresentation: Record<
  TransactionStatus,
  StatusPresentation
> = {
  awaiting_signature: { label: "Awaiting signature", variant: "info" },
  pending: { label: "Pending", variant: "warning" },
  confirmed: { label: "Confirmed", variant: "success" },
  failed: { label: "Failed", variant: "danger" },
};

export type TrustTier = "new" | "verified" | "trusted";

export const trustTierPresentation: Record<TrustTier, StatusPresentation> = {
  new: { label: "New", variant: "neutral" },
  verified: { label: "Verified", variant: "info" },
  trusted: { label: "Trusted", variant: "primary" },
};

export type VerificationStatus = "verified" | "unverified" | "rejected";

export const verificationStatusPresentation: Record<
  VerificationStatus,
  StatusPresentation
> = {
  verified: { label: "Verified", variant: "success" },
  unverified: { label: "Unverified", variant: "warning" },
  rejected: { label: "Rejected", variant: "danger" },
};

export type ProvenanceBand = "complete" | "strong" | "partial" | "limited";

export const provenanceBandPresentation: Record<
  ProvenanceBand,
  StatusPresentation
> = {
  complete: { label: "Complete", variant: "success" },
  strong: { label: "Strong", variant: "primary" },
  partial: { label: "Partial", variant: "warning" },
  limited: { label: "Limited", variant: "neutral" },
};

export const provenanceBandForScore = (score: number): ProvenanceBand => {
  if (score >= 90) return "complete";
  if (score >= 70) return "strong";
  if (score >= 40) return "partial";
  return "limited";
};
