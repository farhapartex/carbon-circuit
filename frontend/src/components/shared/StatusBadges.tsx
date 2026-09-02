import { ShieldCheck } from "lucide-react";
import { StatusPill } from "@/components/shared/StatusPill";
import {
  claimStatusPresentation,
  facilityVerificationPresentation,
  fraudSeverityPresentation,
  listingStatusPresentation,
  provenanceBandForScore,
  provenanceBandPresentation,
  queuePriorityPresentation,
  transactionStatusPresentation,
  trustTierPresentation,
  verificationStatusPresentation,
  type ClaimStatus,
  type FacilityVerificationStatus,
  type FraudSeverity,
  type ListingStatus,
  type QueuePriority,
  type TransactionStatus,
  type TrustTier,
  type VerificationStatus,
} from "@/lib/status";
import type { ProductCategory } from "@/lib/types";
import { cn } from "@/lib/utils";

export const ClaimStatusPill = ({ status }: { status: ClaimStatus }) => (
  <StatusPill presentation={claimStatusPresentation[status]} />
);

export const ListingStatusPill = ({ status }: { status: ListingStatus }) => (
  <StatusPill presentation={listingStatusPresentation[status]} />
);

export const TransactionStatusPill = ({
  status,
}: {
  status: TransactionStatus;
}) => <StatusPill presentation={transactionStatusPresentation[status]} />;

export const TrustTierBadge = ({ tier }: { tier: TrustTier }) => (
  <StatusPill presentation={trustTierPresentation[tier]} />
);

export const FacilityVerificationBadge = ({
  status,
}: {
  status: FacilityVerificationStatus;
}) => <StatusPill presentation={facilityVerificationPresentation[status]} />;

export const VerificationStatusBadge = ({
  status,
}: {
  status: VerificationStatus;
}) => <StatusPill presentation={verificationStatusPresentation[status]} />;

export const PriorityBadge = ({ priority }: { priority: QueuePriority }) => (
  <StatusPill presentation={queuePriorityPresentation[priority]} />
);

export const FraudSeverityBadge = ({
  severity,
}: {
  severity: FraudSeverity;
}) => <StatusPill presentation={fraudSeverityPresentation[severity]} />;

type ProvenanceScoreBadgeProps = {
  score: number;
  showScore?: boolean;
  className?: string;
};

export function ProvenanceScoreBadge({
  score,
  showScore = true,
  className,
}: ProvenanceScoreBadgeProps) {
  const band = provenanceBandForScore(score);
  const presentation = provenanceBandPresentation[band];

  return (
    <StatusPill
      presentation={
        showScore
          ? { ...presentation, label: `${presentation.label} · ${score}` }
          : presentation
      }
      className={className}
    />
  );
}

const PRODUCT_CATEGORY_LABELS: Record<ProductCategory, string> = {
  electronics: "Electronics",
  agriculture: "Agriculture",
  pharma: "Pharma",
  textiles: "Textiles",
};

export const ProductCategoryBadge = ({
  category,
}: {
  category: ProductCategory;
}) => (
  <StatusPill
    presentation={{
      label: PRODUCT_CATEGORY_LABELS[category],
      variant: "neutral",
    }}
    showDot={false}
  />
);

export function ExpiryWarningBadge({
  daysRemaining,
}: {
  daysRemaining: number;
}) {
  return (
    <StatusPill
      presentation={{
        label:
          daysRemaining <= 0
            ? "Expired"
            : `Expires in ${daysRemaining} day${daysRemaining === 1 ? "" : "s"}`,
        variant: daysRemaining <= 0 ? "neutral" : "warning",
      }}
    />
  );
}

export function MFARequiredBadge({ className }: { className?: string }) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 text-caption text-muted-foreground",
        className,
      )}
    >
      <ShieldCheck className="size-3.5 text-primary-600" aria-hidden />
      MFA required
    </span>
  );
}
