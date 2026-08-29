import type { Metadata } from "next";
import { PackageSearch } from "lucide-react";
import { notFound } from "next/navigation";
import { CheckpointTimeline } from "@/components/features/provenance/CheckpointTimeline";
import { ProvenanceScorePanel } from "@/components/features/provenance/ProvenanceScorePanel";
import { SustainabilityHighlightCard } from "@/components/features/provenance/SustainabilityHighlightCard";
import { VerifyOnChainLink } from "@/components/features/provenance/VerifyOnChainLink";
import { ProductCategoryBadge } from "@/components/shared/StatusBadges";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { getPublicBatchView } from "@/lib/fixtures";

export const metadata: Metadata = {
  title: "Batch provenance",
  description: "The recorded journey of this batch, open to anyone.",
};

export default async function TrackBatchPage(
  props: PageProps<"/track/[publicRef]">,
) {
  const { publicRef } = await props.params;
  const batch = await getPublicBatchView(publicRef);

  if (!batch) notFound();

  return (
    <article className="space-y-6">
      <header className="space-y-3">
        <div className="flex flex-wrap items-center gap-3">
          <ProductCategoryBadge category={batch.productCategory} />
          <span className="text-caption text-neutral-600">
            Last updated <TimestampDisplay value={batch.lastUpdatedAt} />
          </span>
        </div>
        <h1 className="text-page-title text-balance">{batch.componentType}</h1>
        <p className="text-body text-neutral-600">
          Produced by {batch.originatingFacilityName} in{" "}
          {batch.originatingFacilityCountry} on{" "}
          <TimestampDisplay value={batch.producedAt} dateOnly />.
        </p>
      </header>

      <ProvenanceScorePanel score={batch.provenanceScore} />

      <SustainabilityHighlightCard
        facilityName={batch.originatingFacilityName}
        claims={batch.approvedClaimSummaries}
      />

      <section className="space-y-4 rounded-lg border border-neutral-200 bg-white p-6">
        <h2 className="font-medium">Recorded journey</h2>
        {batch.checkpoints.length === 0 ? (
          <div className="flex flex-col items-center gap-3 py-10 text-center">
            <PackageSearch
              className="size-6 text-muted-foreground"
              aria-hidden
            />
            <p className="font-medium">
              This batch has no recorded journey yet
            </p>
            <p className="max-w-md text-caption text-pretty text-neutral-600">
              It has been registered, but nobody has logged a checkpoint for it
              so far. That is why the score is low. It does not mean anything is
              wrong with the product.
            </p>
          </div>
        ) : (
          <CheckpointTimeline checkpoints={batch.checkpoints} />
        )}
      </section>

      <VerifyOnChainLink checkpoints={batch.checkpoints} />

      <p className="text-caption text-pretty text-neutral-600">
        This page shows what was recorded about the batch&apos;s journey. It
        deliberately does not show quantities, prices, or the names of companies
        that handled it, because that is commercially sensitive information
        nobody agreed to publish by joining a traceability platform.
      </p>
    </article>
  );
}
