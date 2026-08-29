import { PageHeader } from "@/components/shared/PageHeader";

export default async function BatchesBatchidCheckpointsNewPage(
  props: PageProps<"/batches/[batchId]/checkpoints/new">,
) {
  const { batchId } = await props.params;
  return <PageHeader title="Log a checkpoint" description={batchId} />;
}
