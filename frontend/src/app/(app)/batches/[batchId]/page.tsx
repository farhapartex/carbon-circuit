import { PageHeader } from "@/components/shared/PageHeader";

export default async function BatchesBatchidPage(
  props: PageProps<"/batches/[batchId]">,
) {
  const { batchId } = await props.params;
  return <PageHeader title="Batch" description={batchId} />;
}
