import type { Metadata } from "next";
import { ListingForm } from "@/components/features/marketplace/ListingForm";
import { PageHeader } from "@/components/shared/PageHeader";
import { fetchPlans } from "@/lib/api/plans";
import { getCreditPortfolio, getSubscription } from "@/lib/fixtures";

export const metadata: Metadata = { title: "Create a listing" };

export default async function NewListingPage() {
  const [portfolio, subscription, plans] = await Promise.all([
    getCreditPortfolio(),
    getSubscription(),
    fetchPlans(),
  ]);

  const plan = plans.find(
    (candidate) => candidate.tier === subscription.planTier,
  );

  return (
    <>
      <PageHeader
        backTo={{ href: "/marketplace/my-listings", label: "My listings" }}
        title="Create a listing"
        description="Sell credits from one credit class at a price per tonne in USDC."
      />

      <ListingForm
        balances={portfolio.balances}
        feeBasisPoints={plan?.marketplaceFeeBasisPoints ?? null}
      />
    </>
  );
}
