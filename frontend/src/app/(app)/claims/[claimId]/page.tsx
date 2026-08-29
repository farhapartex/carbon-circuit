import { PageHeader } from "@/components/shared/PageHeader";

export default async function ClaimsClaimidPage(
  props: PageProps<"/claims/[claimId]">,
) {
  const { claimId } = await props.params;
  return <PageHeader title="Claim" description={claimId} />;
}
