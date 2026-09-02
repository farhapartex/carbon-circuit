import type { Metadata } from "next";
import Link from "next/link";
import { ListingBrowser } from "@/components/features/marketplace/ListingBrowser";
import { PageHeader } from "@/components/shared/PageHeader";
import { Button } from "@/components/ui/button";
import { listListings } from "@/lib/fixtures";

export const metadata: Metadata = { title: "Marketplace" };

export default async function MarketplacePage() {
  const listings = await listListings();

  return (
    <>
      <PageHeader
        title="Marketplace"
        description="Credits are sold by credit class, so you always know which facility's verified practice you are buying."
        actions={
          <Button asChild variant="outline">
            <Link href="/marketplace/my-listings">My listings</Link>
          </Button>
        }
      />

      <ListingBrowser listings={listings.items} />
    </>
  );
}
