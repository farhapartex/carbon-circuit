import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { DoubleCountingDisclosure } from "@/components/features/marketplace/DoubleCountingDisclosure";
import { PurchaseForm } from "@/components/features/marketplace/PurchaseForm";
import { SellerFacilityCard } from "@/components/features/marketplace/SellerFacilityCard";
import { CreditAmountDisplay } from "@/components/shared/CreditAmountDisplay";
import { PageHeader } from "@/components/shared/PageHeader";
import { ListingStatusPill } from "@/components/shared/StatusBadges";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { UsdcAmountDisplay } from "@/components/shared/UsdcAmountDisplay";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { fetchCurrentOrganization } from "@/lib/api/organization";
import { auth0 } from "@/lib/auth0";
import { getListing } from "@/lib/fixtures";
import { activityTypeLabels } from "@/lib/labels";

export const metadata: Metadata = { title: "Listing" };

export default async function ListingDetailPage({
  params,
}: PageProps<"/marketplace/listings/[listingId]">) {
  const { listingId } = await params;
  const listing = await getListing(listingId);

  if (!listing) notFound();

  const { token } = await auth0.getAccessToken();
  const organization = await fetchCurrentOrganization(token);

  return (
    <>
      <PageHeader
        backTo={{ href: "/marketplace", label: "Marketplace" }}
        title={`${activityTypeLabels[listing.creditClass.activityType]}, vintage ${listing.creditClass.vintageYear}`}
        description={`Listed by ${listing.sellerOrganizationName}.`}
        meta={<ListingStatusPill status={listing.status} />}
      />

      <div className="grid gap-6 lg:grid-cols-[3fr_2fr]">
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Listing</CardTitle>
            </CardHeader>
            <CardContent>
              <dl className="space-y-3">
                <div className="flex flex-wrap items-baseline justify-between gap-4">
                  <dt className="text-caption text-neutral-600">
                    Quantity remaining
                  </dt>
                  <dd className="font-medium">
                    <CreditAmountDisplay amount={listing.quantityRemaining} />
                  </dd>
                </div>
                <div className="flex flex-wrap items-baseline justify-between gap-4">
                  <dt className="text-caption text-neutral-600">
                    Originally listed
                  </dt>
                  <dd className="font-medium">
                    <CreditAmountDisplay amount={listing.quantityOriginal} />
                  </dd>
                </div>
                <div className="flex flex-wrap items-baseline justify-between gap-4">
                  <dt className="text-caption text-neutral-600">
                    Price per tCO2e
                  </dt>
                  <dd className="font-medium">
                    <UsdcAmountDisplay amount={listing.pricePerTonne} />
                  </dd>
                </div>
                <div className="flex flex-wrap items-baseline justify-between gap-4">
                  <dt className="text-caption text-neutral-600">
                    Minimum purchase
                  </dt>
                  <dd className="font-medium">
                    <CreditAmountDisplay
                      amount={listing.minimumPurchaseQuantity}
                    />
                  </dd>
                </div>
                <div className="flex flex-wrap items-baseline justify-between gap-4">
                  <dt className="text-caption text-neutral-600">Expires</dt>
                  <dd className="font-medium">
                    <TimestampDisplay value={listing.expiresAt} dateOnly />
                  </dd>
                </div>
              </dl>
            </CardContent>
          </Card>

          <SellerFacilityCard listing={listing} />

          <DoubleCountingDisclosure />
        </div>

        <div className="space-y-6">
          <PurchaseForm
            listing={listing}
            treasuryAddress={organization.treasuryAddress}
          />
        </div>
      </div>
    </>
  );
}
