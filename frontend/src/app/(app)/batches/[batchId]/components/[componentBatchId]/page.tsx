import type { Metadata, Route } from "next";
import { notFound } from "next/navigation";
import { CheckpointTimeline } from "@/components/features/provenance/CheckpointTimeline";
import { ProvenanceScorePanel } from "@/components/features/provenance/ProvenanceScorePanel";
import { PageHeader } from "@/components/shared/PageHeader";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { fetchBatch, fetchComponentBatch } from "@/lib/api/batches";
import { GatewayError } from "@/lib/api/gateway";
import { auth0 } from "@/lib/auth0";
import { productCategoryLabels } from "@/lib/labels";

export const metadata: Metadata = { title: "Component batch" };

const missingIsNull = (error: unknown) => {
  if (error instanceof GatewayError && error.status === 404) return null;
  throw error;
};

export default async function ComponentBatchPage({
  params,
}: PageProps<"/batches/[batchId]/components/[componentBatchId]">) {
  const { batchId, componentBatchId } = await params;
  const { token } = await auth0.getAccessToken();

  const [parent, component] = await Promise.all([
    fetchBatch(token, batchId).catch(missingIsNull),
    fetchComponentBatch(token, batchId, componentBatchId).catch(missingIsNull),
  ]);

  if (!parent || !component) notFound();

  return (
    <>
      <PageHeader
        backTo={{
          href: `/batches/${parent.batch.id}` as Route,
          label: parent.batch.componentType,
        }}
        title={component.batch.componentType}
        description={`A component batch that ${parent.batch.componentType} descends from, produced at ${component.batch.originatingFacilityName}.`}
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
        </div>

        <div className="space-y-6">
          <ProvenanceScorePanel score={component.batch.provenanceScore} />

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
                    {component.batch.originatingFacilityName}
                  </dd>
                </div>
                <div>
                  <dt className="text-caption text-neutral-600">
                    Product category
                  </dt>
                  <dd className="font-medium">
                    {productCategoryLabels[component.batch.productCategory]}
                  </dd>
                </div>
                <div>
                  <dt className="text-caption text-neutral-600">Produced</dt>
                  <dd className="font-medium">
                    <TimestampDisplay
                      value={component.batch.producedAt}
                      dateOnly
                    />
                  </dd>
                </div>
              </dl>
            </CardContent>
          </Card>
        </div>
      </div>

      <p className="max-w-2xl text-caption text-pretty text-neutral-600">
        You can see this component batch because {parent.batch.componentType}{" "}
        declares it as a parent. Provenance is visible one level up only —
        quantities, the companies that handled it, and its own suppliers stay
        with the organization that owns it, and that limit is enforced by the
        database rather than by this page.
      </p>
    </>
  );
}
