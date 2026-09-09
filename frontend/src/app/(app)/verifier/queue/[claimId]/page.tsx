import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { CeilingComparisonCard } from "@/components/features/claims/CeilingComparisonCard";
import { ClaimDetailsCard } from "@/components/features/claims/ClaimDetailsCard";
import { EvidenceViewer } from "@/components/features/claims/EvidenceViewer";
import { AIAssessmentPanel } from "@/components/features/verifier/AIAssessmentPanel";
import { DecisionActionBar } from "@/components/features/verifier/DecisionActionBar";
import { DecisionHistory } from "@/components/features/verifier/DecisionHistory";
import { PageHeader } from "@/components/shared/PageHeader";
import {
  ClaimStatusPill,
  PriorityBadge,
} from "@/components/shared/StatusBadges";
import { toSustainabilityClaim } from "@/lib/api/claimView";
import { GatewayError } from "@/lib/api/gateway";
import { fetchReview } from "@/lib/api/verifier";
import { auth0 } from "@/lib/auth0";
import { activityTypeLabels } from "@/lib/labels";

export const metadata: Metadata = { title: "Review claim" };

export default async function VerifierReviewPage({
  params,
}: PageProps<"/verifier/queue/[claimId]">) {
  const { claimId } = await params;
  const { token } = await auth0.getAccessToken();

  const review = await fetchReview(token, claimId).catch((error: unknown) => {
    if (error instanceof GatewayError && error.status === 404) return null;
    throw error;
  });

  if (!review) notFound();

  const claim = toSustainabilityClaim(review.claim);
  const approvals = review.decisions.filter(
    (decision) => decision.outcome === "approved",
  ).length;

  const open = review.claim.status === "under_human_review";

  return (
    <>
      <PageHeader
        backTo={{ href: "/verifier/queue", label: "Review queue" }}
        title={`${activityTypeLabels[claim.activityType]}, vintage ${claim.vintageYear}`}
        description={`Filed for ${claim.facilityName}.`}
        meta={
          <>
            <ClaimStatusPill status={claim.status} />
            <PriorityBadge priority={claim.priority} />
          </>
        }
      />

      <div className="grid gap-6 lg:grid-cols-[3fr_2fr]">
        <div className="space-y-6">
          <ClaimDetailsCard claim={claim} />
          <AIAssessmentPanel review={review.aiReview} />
          <EvidenceViewer evidence={claim.evidence} />
        </div>

        <div className="space-y-6">
          <CeilingComparisonCard claim={claim} />
          <DecisionHistory
            decisions={review.decisions}
            requiresDualApproval={review.claim.requiresDualApproval}
          />
          {open ? (
            <DecisionActionBar
              claimId={review.claim.id}
              requestedAmount={review.claim.requestedAmount}
              computedCeiling={review.claim.computedCeiling}
              requiresDualApproval={review.claim.requiresDualApproval}
              approvalsSoFar={approvals}
            />
          ) : (
            <div className="rounded-lg border border-neutral-200 bg-white px-4 py-4">
              <p className="font-medium">This claim is settled</p>
              <p className="mt-1 text-caption text-pretty text-neutral-600">
                A recorded decision is final. Correcting one is a separate,
                logged action taken by an admin, never an edit.
              </p>
            </div>
          )}
        </div>
      </div>
    </>
  );
}
