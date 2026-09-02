"use client";

import { Store } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { CreditAmountDisplay } from "@/components/shared/CreditAmountDisplay";
import { DataTable, type DataTableColumn } from "@/components/shared/DataTable";
import { EmptyState } from "@/components/shared/EmptyState";
import {
  ExpiryWarningBadge,
  ListingStatusPill,
} from "@/components/shared/StatusBadges";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { UsdcAmountDisplay } from "@/components/shared/UsdcAmountDisplay";
import { Button } from "@/components/ui/button";
import { activityTypeLabels } from "@/lib/labels";
import type { MarketplaceListing } from "@/lib/types";

const daysUntil = (iso: string) =>
  Math.ceil((new Date(iso).getTime() - Date.now()) / 86_400_000);

const cancellable = (listing: MarketplaceListing) =>
  listing.status === "active" || listing.status === "partially_filled";

export function MyListingsTable({
  listings,
}: {
  listings: MarketplaceListing[];
}) {
  const [cancelling, setCancelling] = useState<MarketplaceListing | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  const columns: DataTableColumn<MarketplaceListing>[] = [
    {
      key: "class",
      header: "Credit class",
      render: (listing) => (
        <span className="block">
          <span className="block font-medium">
            {activityTypeLabels[listing.creditClass.activityType]}
          </span>
          <span className="block text-caption text-neutral-600">
            {listing.creditClass.facilityName} · vintage{" "}
            {listing.creditClass.vintageYear}
          </span>
        </span>
      ),
    },
    {
      key: "remaining",
      header: "Remaining",
      alignEnd: true,
      render: (listing) => (
        <CreditAmountDisplay amount={listing.quantityRemaining} />
      ),
    },
    {
      key: "price",
      header: "Price",
      alignEnd: true,
      render: (listing) => <UsdcAmountDisplay amount={listing.pricePerTonne} />,
    },
    {
      key: "status",
      header: "Status",
      render: (listing) => <ListingStatusPill status={listing.status} />,
    },
    {
      key: "expiry",
      header: "Expires",
      render: (listing) => {
        const remaining = daysUntil(listing.expiresAt);

        return remaining <= 7 && cancellable(listing) ? (
          <ExpiryWarningBadge daysRemaining={remaining} />
        ) : (
          <TimestampDisplay value={listing.expiresAt} dateOnly />
        );
      },
    },
    {
      key: "actions",
      header: "Actions",
      alignEnd: true,
      render: (listing) =>
        cancellable(listing) ? (
          <Button
            size="sm"
            variant="outline"
            onClick={() => setCancelling(listing)}
          >
            Cancel
          </Button>
        ) : null,
    },
  ];

  return (
    <div className="space-y-4">
      {notice ? (
        <div
          role="alert"
          className="rounded-md border border-warning-600 bg-warning-50 px-4 py-3"
        >
          <p className="text-helper text-warning-700">{notice}</p>
        </div>
      ) : null}

      <DataTable
        columns={columns}
        rows={listings}
        rowKey={(listing) => listing.id}
        caption="Listings created by your organization"
        emptyState={
          <EmptyState
            icon={Store}
            title="No listings yet"
            description="Listing credits moves them into escrow immediately, which is what makes overselling impossible rather than merely checked for."
            action={
              <Button asChild>
                <Link href="/marketplace/my-listings/new">
                  Create a listing
                </Link>
              </Button>
            }
          />
        }
      />

      <ConfirmDialog
        open={cancelling !== null}
        onOpenChange={(open) => {
          if (!open) setCancelling(null);
        }}
        title="Cancel this listing?"
        description="The credits still held in escrow against this listing return to your available balance. Any quantity already sold is unaffected."
        confirmLabel="Cancel listing"
        destructive
        consequence={
          cancelling ? (
            <span>
              <CreditAmountDisplay amount={cancelling.quantityRemaining} />{" "}
              returns to your balance.
            </span>
          ) : undefined
        }
        onConfirm={() => {
          setCancelling(null);
          setNotice(
            "Listings cannot be cancelled yet — the marketplace service is not built.",
          );
        }}
      />
    </div>
  );
}
