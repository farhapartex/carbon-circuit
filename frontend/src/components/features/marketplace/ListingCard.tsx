import Link from "next/link";
import { CreditAmountDisplay } from "@/components/shared/CreditAmountDisplay";
import {
  ExpiryWarningBadge,
  TrustTierBadge,
} from "@/components/shared/StatusBadges";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { UsdcAmountDisplay } from "@/components/shared/UsdcAmountDisplay";
import { Button } from "@/components/ui/button";
import { countryName } from "@/lib/countries";
import { activityTypeLabels } from "@/lib/labels";
import type { MarketplaceListing } from "@/lib/types";

const daysUntil = (iso: string) =>
  Math.ceil((new Date(iso).getTime() - Date.now()) / 86_400_000);

export function ListingCard({ listing }: { listing: MarketplaceListing }) {
  const remaining = daysUntil(listing.expiresAt);

  return (
    <article className="flex flex-col gap-4 rounded-lg border border-neutral-200 bg-white p-5 shadow-sm">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h3 className="font-medium">
            {activityTypeLabels[listing.creditClass.activityType]}, vintage{" "}
            {listing.creditClass.vintageYear}
          </h3>
          <p className="text-caption text-neutral-600">
            {listing.creditClass.facilityName} ·{" "}
            {countryName(listing.creditClass.facilityCountry)}
          </p>
        </div>
        {remaining <= 7 ? (
          <ExpiryWarningBadge daysRemaining={remaining} />
        ) : null}
      </div>

      <dl className="grid gap-3 sm:grid-cols-3">
        <div>
          <dt className="text-caption text-neutral-600">Available</dt>
          <dd className="font-medium">
            <CreditAmountDisplay amount={listing.quantityRemaining} />
          </dd>
        </div>
        <div>
          <dt className="text-caption text-neutral-600">Price per tCO2e</dt>
          <dd className="font-medium">
            <UsdcAmountDisplay amount={listing.pricePerTonne} />
          </dd>
        </div>
        <div>
          <dt className="text-caption text-neutral-600">Minimum purchase</dt>
          <dd className="font-medium">
            <CreditAmountDisplay amount={listing.minimumPurchaseQuantity} />
          </dd>
        </div>
      </dl>

      <div className="flex flex-wrap items-center justify-between gap-3 border-t border-neutral-200 pt-4">
        <span className="flex flex-wrap items-center gap-2">
          <span className="text-caption text-neutral-600">
            {listing.sellerOrganizationName}
          </span>
          <TrustTierBadge tier={listing.sellerTrustTier} />
        </span>
        <span className="flex items-center gap-3">
          <span className="text-caption text-neutral-600">
            Expires <TimestampDisplay value={listing.expiresAt} dateOnly />
          </span>
          <Button asChild size="sm">
            <Link href={`/marketplace/listings/${listing.id}`}>View</Link>
          </Button>
        </span>
      </div>
    </article>
  );
}
