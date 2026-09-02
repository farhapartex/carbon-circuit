"use client";

import { Boxes, FileCheck2, Coins } from "lucide-react";
import Link from "next/link";
import { CapacityCard } from "@/components/features/facilities/CapacityCard";
import { FacilityProfileCard } from "@/components/features/facilities/FacilityProfileCard";
import { CreditAmountDisplay } from "@/components/shared/CreditAmountDisplay";
import { DataTable, type DataTableColumn } from "@/components/shared/DataTable";
import { EmptyState } from "@/components/shared/EmptyState";
import { MetricCard } from "@/components/shared/MetricCard";
import {
  ClaimStatusPill,
  ProvenanceScoreBadge,
} from "@/components/shared/StatusBadges";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { activityTypeLabels } from "@/lib/labels";
import type {
  Batch,
  CreditClassBalance,
  Facility,
  SustainabilityClaim,
} from "@/lib/types";

const numberFormat = new Intl.NumberFormat("en-US");

const batchColumns: DataTableColumn<Batch>[] = [
  {
    key: "component",
    header: "Batch",
    render: (batch) => (
      <Link
        href={`/batches/${batch.id}`}
        className="font-medium underline-offset-4 hover:underline"
      >
        {batch.componentType}
      </Link>
    ),
  },
  {
    key: "quantity",
    header: "Quantity",
    alignEnd: true,
    render: (batch) => (
      <span className="tabular-nums">
        {numberFormat.format(batch.quantity)} {batch.unit}
      </span>
    ),
  },
  {
    key: "score",
    header: "Provenance",
    alignEnd: true,
    render: (batch) => (
      <ProvenanceScoreBadge score={batch.provenanceScore.total} />
    ),
  },
  {
    key: "produced",
    header: "Produced",
    hideOnCard: true,
    render: (batch) => <TimestampDisplay value={batch.producedAt} dateOnly />,
  },
];

const claimColumns: DataTableColumn<SustainabilityClaim>[] = [
  {
    key: "activity",
    header: "Activity",
    render: (claim) => (
      <Link
        href={`/claims/${claim.id}`}
        className="font-medium underline-offset-4 hover:underline"
      >
        {activityTypeLabels[claim.activityType]}
      </Link>
    ),
  },
  {
    key: "vintage",
    header: "Vintage",
    render: (claim) => (
      <span className="tabular-nums">{claim.vintageYear}</span>
    ),
  },
  {
    key: "requested",
    header: "Requested",
    alignEnd: true,
    render: (claim) => <CreditAmountDisplay amount={claim.requestedAmount} />,
  },
  {
    key: "status",
    header: "Status",
    render: (claim) => <ClaimStatusPill status={claim.status} />,
  },
];

const balanceColumns: DataTableColumn<CreditClassBalance>[] = [
  {
    key: "vintage",
    header: "Credit class",
    render: (balance) => (
      <span className="block">
        <span className="block font-medium">
          {activityTypeLabels[balance.creditClass.activityType]}
        </span>
        <span className="block text-caption text-neutral-600">
          Vintage {balance.creditClass.vintageYear}
        </span>
      </span>
    ),
  },
  {
    key: "available",
    header: "Available",
    alignEnd: true,
    render: (balance) => <CreditAmountDisplay amount={balance.available} />,
  },
  {
    key: "escrowed",
    header: "Escrowed",
    alignEnd: true,
    render: (balance) => <CreditAmountDisplay amount={balance.escrowed} />,
  },
  {
    key: "retired",
    header: "Retired",
    alignEnd: true,
    render: (balance) => <CreditAmountDisplay amount={balance.retired} />,
  },
];

type FacilityTabsProps = {
  facility: Facility;
  batches: Batch[];
  claims: SustainabilityClaim[];
  balances: CreditClassBalance[];
};

export function FacilityTabs({
  facility,
  batches,
  claims,
  balances,
}: FacilityTabsProps) {
  return (
    <Tabs defaultValue="overview">
      <TabsList>
        <TabsTrigger value="overview">Overview</TabsTrigger>
        <TabsTrigger value="batches">Batches</TabsTrigger>
        <TabsTrigger value="claims">Claims</TabsTrigger>
        <TabsTrigger value="credits">Credits</TabsTrigger>
      </TabsList>

      <TabsContent value="overview" className="space-y-6">
        <div className="grid gap-4 sm:grid-cols-3">
          <MetricCard
            label="Batches produced"
            value={numberFormat.format(facility.batchCount)}
          />
          <MetricCard
            label="Approved claims"
            value={numberFormat.format(facility.approvedClaimCount)}
          />
          <MetricCard
            label="Ceiling discount"
            value={facility.ceilingDiscountFactor}
            hint="Applied to every credit ceiling computed for this facility."
          />
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          <FacilityProfileCard facility={facility} />
          <CapacityCard facility={facility} />
        </div>
      </TabsContent>

      <TabsContent value="batches">
        <DataTable
          columns={batchColumns}
          rows={batches}
          rowKey={(batch) => batch.id}
          caption={`Batches originating at ${facility.name}`}
          emptyState={
            <EmptyState
              icon={Boxes}
              title="No batches from this facility"
              description="Batches registered against this site will appear here."
              action={
                <Button asChild>
                  <Link href="/batches/new">Register a batch</Link>
                </Button>
              }
            />
          }
        />
      </TabsContent>

      <TabsContent value="claims">
        <DataTable
          columns={claimColumns}
          rows={claims}
          rowKey={(claim) => claim.id}
          caption={`Sustainability claims filed for ${facility.name}`}
          emptyState={
            <EmptyState
              icon={FileCheck2}
              title="No claims for this facility"
              description="A sustainability claim converts this facility's verified activity into carbon credits."
              action={
                <Button asChild>
                  <Link href="/claims/new">Submit a claim</Link>
                </Button>
              }
            />
          }
        />
      </TabsContent>

      <TabsContent value="credits">
        <DataTable
          columns={balanceColumns}
          rows={balances}
          rowKey={(balance) => balance.creditClass.tokenId}
          caption={`Credits issued from ${facility.name}`}
          emptyState={
            <EmptyState
              icon={Coins}
              title="No credits issued yet"
              description="Credits appear here once a verifier approves a claim for this facility."
              action={
                <Button asChild variant="outline">
                  <Link href="/credits">View all credits</Link>
                </Button>
              }
            />
          }
        />
      </TabsContent>
    </Tabs>
  );
}
