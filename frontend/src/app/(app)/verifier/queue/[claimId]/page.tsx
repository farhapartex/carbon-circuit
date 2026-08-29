import { PageHeader } from "@/components/shared/PageHeader";

export default async function VerifierQueueClaimidPage(
  props: PageProps<"/verifier/queue/[claimId]">,
) {
  const { claimId } = await props.params;
  return <PageHeader title="Review claim" description={claimId} />;
}
