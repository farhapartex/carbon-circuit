import type { Metadata, Route } from "next";
import { notFound } from "next/navigation";
import { CheckpointForm } from "@/components/features/provenance/CheckpointForm";
import { PageHeader } from "@/components/shared/PageHeader";
import { getBatch } from "@/lib/fixtures";

export const metadata: Metadata = { title: "Log a checkpoint" };

export default async function LogCheckpointPage({
  params,
}: PageProps<"/batches/[batchId]/checkpoints/new">) {
  const { batchId } = await params;
  const batch = await getBatch(batchId);

  if (!batch) notFound();

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
