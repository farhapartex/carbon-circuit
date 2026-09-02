"use client";

import { Factory } from "lucide-react";
import Link from "next/link";
import { DataTable, type DataTableColumn } from "@/components/shared/DataTable";
import { EmptyState } from "@/components/shared/EmptyState";
import {
  FacilityVerificationBadge,
  TrustTierBadge,
} from "@/components/shared/StatusBadges";
import { Button } from "@/components/ui/button";
import { countryName } from "@/lib/countries";
import { facilityTypeLabels } from "@/lib/labels";
import type { Facility } from "@/lib/types";

const numberFormat = new Intl.NumberFormat("en-US");

const columns: DataTableColumn<Facility>[] = [
  {
    key: "name",
    header: "Facility",
    render: (facility) => (
      <span className="block">
        <Link
          href={`/facilities/${facility.id}`}
          className="font-medium underline-offset-4 hover:underline"
        >
          {facility.name}
        </Link>
        <span className="block text-caption text-neutral-600">
          {countryName(facility.countryCode)}
        </span>
      </span>
    ),
  },
  {
    key: "type",
    header: "Type",
    render: (facility) => facilityTypeLabels[facility.type],
  },
  {
    key: "gridRegion",
    header: "Grid region",
    render: (facility) => (
      <span className="font-mono text-helper">{facility.gridRegion}</span>
    ),
  },
  {
    key: "verification",
    header: "Verification",
    render: (facility) => (
      <FacilityVerificationBadge status={facility.verificationStatus} />
    ),
  },
  {
    key: "trustTier",
    header: "Trust tier",
    render: (facility) => <TrustTierBadge tier={facility.trustTier} />,
  },
  {
    key: "batches",
    header: "Batches",
    alignEnd: true,
    render: (facility) => (
      <span className="tabular-nums">
        {numberFormat.format(facility.batchCount)}
      </span>
    ),
  },
];

export function FacilityTable({ facilities }: { facilities: Facility[] }) {
  return (
    <DataTable
      columns={columns}
      rows={facilities}
      rowKey={(facility) => facility.id}
      caption="Facilities registered under your organization"
      emptyState={
        <EmptyState
          icon={Factory}
          title="No facilities yet"
          description="A facility is a physical site that produces batches. Its registry match determines how much of its declared capacity can back a credit claim."
          action={
            <Button asChild>
              <Link href="/facilities/new">Add a facility</Link>
            </Button>
          }
        />
      }
    />
  );
}
