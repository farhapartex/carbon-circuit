import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { CheckpointTimeline } from "@/components/features/provenance/CheckpointTimeline";
import { ParentBatchLinks } from "@/components/features/provenance/ParentBatchLinks";
import { ProvenanceScorePanel } from "@/components/features/provenance/ProvenanceScorePanel";
import { CopyButton } from "@/components/shared/CopyButton";
import { PageHeader } from "@/components/shared/PageHeader";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { fetchBatch, fetchCheckpoints } from "@/lib/api/batches";
import { GatewayError } from "@/lib/api/gateway";
import { auth0 } from "@/lib/auth0";
import { formatQuantity } from "@/lib/decimal";
import { productCategoryLabels } from "@/lib/labels";

export const metadata: Metadata = { title: "Batch" };

export default async function BatchDetailPage({
  params,
}: PageProps<"/batches/[batchId]">) {
  const { batchId } = await params;
  const { token } = await auth0.getAccessToken();

  const detail = await fetchBatch(token, batchId).catch((error: unknown) => {
    if (error instanceof GatewayError && error.status === 404) return null;
    throw error;
  });

  if (!detail) notFound();

  const { batch, parents } = detail;
  const checkpoints = await fetchCheckpoints(token, batch.id);

  return (
    <>
      <div className="flex flex-wrap items-start justify-between gap-4">
        <PageHeader
          backTo={{ href: "/batches", label: "Batches" }}
          title={batch.componentType}
          description={`${formatQuantity(batch.quantity)} ${batch.unit} produced at ${batch.originatingFacilityName}.`}
        />
        <Button asChild>
          <Link href={`/batches/${batch.id}/checkpoints/new`}>
            Log a checkpoint
          </Link>
        </Button>
      </div>

      <div className="grid gap-6 lg:grid-cols-[2fr_1fr]">
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Chain of custody</CardTitle>
            </CardHeader>
            <CardContent>
              <CheckpointTimeline checkpoints={checkpoints} showReporter />
            </CardContent>
          </Card>

          <ParentBatchLinks batchId={batch.id} parents={parents} />
        </div>

        <div className="space-y-6">
          <ProvenanceScorePanel score={batch.provenanceScore} showBreakdown />

          <Card>
            <CardHeader>
              <CardTitle>Batch details</CardTitle>
            </CardHeader>
            <CardContent>
              <dl className="space-y-3">
                <div>
                  <dt className="text-caption text-neutral-600">
                    Product category
                  </dt>
                  <dd className="font-medium">
                    {productCategoryLabels[batch.productCategory]}
                  </dd>
                </div>
                {batch.lotNumber ? (
                  <div>
                    <dt className="text-caption text-neutral-600">
                      Lot number
                    </dt>
                    <dd className="font-medium">{batch.lotNumber}</dd>
                  </div>
                ) : null}
                {batch.externalId ? (
                  <div>
                    <dt className="text-caption text-neutral-600">
                      Your reference
                    </dt>
                    <dd className="font-medium">{batch.externalId}</dd>
                  </div>
                ) : null}
                <div>
                  <dt className="text-caption text-neutral-600">Produced</dt>
                  <dd className="font-medium">
                    <TimestampDisplay value={batch.producedAt} dateOnly />
                  </dd>
                </div>
                <div>
                  <dt className="text-caption text-neutral-600">
                    Public reference
                  </dt>
                  <dd className="flex items-center gap-2">
                    <code className="min-w-0 truncate text-helper">
                      {batch.publicReference}
                    </code>
                    <CopyButton
                      value={batch.publicReference}
                      label="Copy public reference"
                    />
                  </dd>
                </div>
              </dl>
            </CardContent>
          </Card>
        </div>
      </div>
    </>
  );
}
