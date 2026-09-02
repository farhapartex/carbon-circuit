import type { Metadata, Route } from "next";
import { notFound } from "next/navigation";
import { CheckpointTimeline } from "@/components/features/provenance/CheckpointTimeline";
import { ProvenanceScorePanel } from "@/components/features/provenance/ProvenanceScorePanel";
import { SustainabilityHighlightCard } from "@/components/features/provenance/SustainabilityHighlightCard";
import { PageHeader } from "@/components/shared/PageHeader";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { countryName } from "@/lib/countries";
import { getBatch, getComponentBatchView } from "@/lib/fixtures";

export const metadata: Metadata = { title: "Component batch" };

export default async function ComponentBatchPage({
  params,
}: PageProps<"/batches/[batchId]/components/[componentBatchId]">) {
  const { batchId, componentBatchId } = await params;

  const [batch, component] = await Promise.all([
    getBatch(batchId),
    getComponentBatchView(batchId, componentBatchId),
  ]);

  if (!batch || !component) notFound();

  return (
    <>
      <PageHeader
        backTo={{
          href: `/batches/${batch.id}` as Route,
          label: batch.componentType,
        }}
        title={component.componentType}
        description={`A component batch that ${batch.componentType} descends from, produced by ${component.originatingFacilityName} in ${countryName(component.originatingFacilityCountry)}.`}
      />

      <div className="grid gap-6 lg:grid-cols-[2fr_1fr]">
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Recorded journey</CardTitle>
            </CardHeader>
            <CardContent>
              <CheckpointTimeline checkpoints={component.checkpoints} />
            </CardContent>
          </Card>

          <SustainabilityHighlightCard
            facilityName={component.originatingFacilityName}
            claims={component.approvedClaimSummaries}
          />
        </div>

        <div className="space-y-6">
          <ProvenanceScorePanel score={component.provenanceScore} />

          <Card>
            <CardHeader>
              <CardTitle>Component details</CardTitle>
            </CardHeader>
            <CardContent>
              <dl className="space-y-3">
                <div>
                  <dt className="text-caption text-neutral-600">
                    Originating facility
                  </dt>
                  <dd className="font-medium">
                    {component.originatingFacilityName}
                  </dd>
                </div>
                <div>
                  <dt className="text-caption text-neutral-600">Country</dt>
                  <dd className="font-medium">
                    {countryName(component.originatingFacilityCountry)}
                  </dd>
                </div>
                <div>
                  <dt className="text-caption text-neutral-600">Produced</dt>
                  <dd className="font-medium">
                    <TimestampDisplay value={component.producedAt} dateOnly />
                  </dd>
                </div>
                <div>
                  <dt className="text-caption text-neutral-600">
                    Last updated
                  </dt>
                  <dd className="font-medium">
                    <TimestampDisplay value={component.lastUpdatedAt} />
                  </dd>
                </div>
              </dl>
            </CardContent>
          </Card>
        </div>
      </div>

      <p className="max-w-2xl text-caption text-pretty text-neutral-600">
        You can see this component batch because {batch.componentType} declares
        it as a parent. Provenance is visible one level up only — quantities,
        the companies that handled it, and its own suppliers stay with the
        organization that owns it.
      </p>
    </>
  );
}
