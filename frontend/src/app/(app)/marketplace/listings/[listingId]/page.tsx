import { PageHeader } from "@/components/shared/PageHeader";

export default async function MarketplaceListingsListingidPage(
  props: PageProps<"/marketplace/listings/[listingId]">,
) {
  const { listingId } = await props.params;
  return <PageHeader title="Listing" description={listingId} />;
}
