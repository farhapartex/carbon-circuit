import { PageHeader } from "@/components/shared/PageHeader";

export default async function TrackPublicrefPage(
  props: PageProps<"/track/[publicRef]">,
) {
  const { publicRef } = await props.params;
  return <PageHeader title="Batch provenance" description={publicRef} />;
}
