import type { Metadata } from "next";
import Link from "next/link";
import { MyListingsTable } from "@/components/features/marketplace/MyListingsTable";
import { PageHeader } from "@/components/shared/PageHeader";
import { Button } from "@/components/ui/button";
import { listListings } from "@/lib/fixtures";

export const metadata: Metadata = { title: "My listings" };

export default async function MyListingsPage() {
  const listings = await listListings();

  return (
    <>
      <PageHeader
        backTo={{ href: "/marketplace", label: "Marketplace" }}
        title="My listings"
        description="Credits on an active listing sit in escrow and are not part of your available balance."
        actions={
          <Button asChild>
            <Link href="/marketplace/my-listings/new">Create a listing</Link>
          </Button>
        }
      />

      <MyListingsTable listings={listings.items} />
    </>
  );
}
