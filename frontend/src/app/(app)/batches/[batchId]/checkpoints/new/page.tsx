import type { Metadata, Route } from "next";
import { notFound } from "next/navigation";
import { CheckpointForm } from "@/components/features/provenance/CheckpointForm";
import { PageHeader } from "@/components/shared/PageHeader";
import { fetchBatch } from "@/lib/api/batches";
import { GatewayError } from "@/lib/api/gateway";
import { auth0 } from "@/lib/auth0";

export const metadata: Metadata = { title: "Log a checkpoint" };

export default async function LogCheckpointPage({
  params,
}: PageProps<"/batches/[batchId]/checkpoints/new">) {
  const { batchId } = await params;
  const { token } = await auth0.getAccessToken();

  const detail = await fetchBatch(token, batchId).catch((error: unknown) => {
    if (error instanceof GatewayError && error.status === 404) return null;
    throw error;
  });

  if (!detail) notFound();

  const { batch } = detail;

  return (
    <>
      <PageHeader
        backTo={{
          href: `/batches/${batch.id}` as Route,
          label: batch.componentType,
        }}
        title="Log a checkpoint"
        description="A checkpoint records where this batch was and when. It is tied to your organization, appended permanently, and forms the chain of custody a consumer sees."
      />

      <CheckpointForm
        batchId={batch.id}
        batchLabel={batch.componentType}
        producedAt={batch.producedAt}
      />
    </>
  );
}
