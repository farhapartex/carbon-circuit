import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { CeilingComparisonCard } from "@/components/features/claims/CeilingComparisonCard";
import { ClaimDetailsCard } from "@/components/features/claims/ClaimDetailsCard";
import { ClaimStatusStepper } from "@/components/features/claims/ClaimStatusStepper";
import { DecisionOutcomeCard } from "@/components/features/claims/DecisionOutcomeCard";
import { EvidenceViewer } from "@/components/features/claims/EvidenceViewer";
import { PageHeader } from "@/components/shared/PageHeader";
import {
  ClaimStatusPill,
  PriorityBadge,
} from "@/components/shared/StatusBadges";
import { getClaim } from "@/lib/fixtures";
import { activityTypeLabels } from "@/lib/labels";

export const metadata: Metadata = { title: "Claim" };

export default async function ClaimDetailPage({
  params,
}: PageProps<"/claims/[claimId]">) {
  const { claimId } = await params;
  const claim = await getClaim(claimId);

  if (!claim) notFound();

  const awaitingResubmission = claim.status === "more_information_requested";

  return (
    <>
      <PageHeader
        backTo={{ href: "/claims", label: "Claims" }}
        title={`${activityTypeLabels[claim.activityType]}, vintage ${claim.vintageYear}`}
        description={`Filed for ${claim.facilityName}.`}
        meta={
          <>
            <ClaimStatusPill status={claim.status} />
            <PriorityBadge priority={claim.priority} />
          </>
        }
      />

      <ClaimStatusStepper status={claim.status} />

      {awaitingResubmission ? (
        <div
          role="status"
          className="rounded-md border border-info-600 bg-info-50 px-4 py-3"
        >
          <p className="font-medium text-info-700">
            A verifier asked for more information
          </p>
          <p className="mt-1 text-caption text-pretty text-info-700">
            Read what they asked for below. Resubmission is not available yet —
            the sustainability service does not exist, and the sitemap defines
            no resubmit route.
          </p>
        </div>
      ) : null}

      <div className="grid gap-6 lg:grid-cols-[3fr_2fr]">
        <div className="space-y-6">
          <ClaimDetailsCard claim={claim} />
          <EvidenceViewer evidence={claim.evidence} />
        </div>

        <div className="space-y-6">
          <CeilingComparisonCard claim={claim} />
          <DecisionOutcomeCard claim={claim} />
        </div>
      </div>
    </>
  );
}
