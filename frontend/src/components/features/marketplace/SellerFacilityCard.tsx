import { TrustTierBadge } from "@/components/shared/StatusBadges";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { countryName } from "@/lib/countries";
import { activityTypeLabels } from "@/lib/labels";
import type { MarketplaceListing } from "@/lib/types";

export function SellerFacilityCard({
  listing,
}: {
  listing: MarketplaceListing;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Who you are buying from</CardTitle>
      </CardHeader>
      <CardContent>
        <dl className="space-y-3">
          <div className="flex flex-wrap items-baseline justify-between gap-4">
            <dt className="text-caption text-neutral-600">Seller</dt>
            <dd className="flex items-center gap-2 font-medium">
              {listing.sellerOrganizationName}
              <TrustTierBadge tier={listing.sellerTrustTier} />
            </dd>
          </div>
          <div className="flex flex-wrap items-baseline justify-between gap-4">
            <dt className="text-caption text-neutral-600">
              Originating facility
            </dt>
            <dd className="font-medium">{listing.creditClass.facilityName}</dd>
          </div>
          <div className="flex flex-wrap items-baseline justify-between gap-4">
            <dt className="text-caption text-neutral-600">Country</dt>
            <dd className="font-medium">
              {countryName(listing.creditClass.facilityCountry)}
            </dd>
          </div>
          <div className="flex flex-wrap items-baseline justify-between gap-4">
            <dt className="text-caption text-neutral-600">Activity</dt>
            <dd className="font-medium">
              {activityTypeLabels[listing.creditClass.activityType]}
            </dd>
          </div>
          <div className="flex flex-wrap items-baseline justify-between gap-4">
            <dt className="text-caption text-neutral-600">Vintage</dt>
            <dd className="font-medium tabular-nums">
              {listing.creditClass.vintageYear}
            </dd>
          </div>
        </dl>

        <p className="mt-4 border-t border-neutral-200 pt-4 text-caption text-pretty text-neutral-600">
          The originating facility is disclosed deliberately rather than
          anonymised, because knowing whose verified practice a credit
          represents is what makes it usable in your own ESG reporting.
        </p>
      </CardContent>
    </Card>
  );
}
