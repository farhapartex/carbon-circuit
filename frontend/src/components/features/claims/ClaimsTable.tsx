"use client";

import { FileCheck2 } from "lucide-react";
import Link from "next/link";
import { useMemo, useState } from "react";
import { CreditAmountDisplay } from "@/components/shared/CreditAmountDisplay";
import { DataTable, type DataTableColumn } from "@/components/shared/DataTable";
import { EmptyState } from "@/components/shared/EmptyState";
import { FilterBar } from "@/components/shared/FilterBar";
import { ClaimStatusPill } from "@/components/shared/StatusBadges";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { activityTypeLabels } from "@/lib/labels";
import { claimStatusPresentation, type ClaimStatus } from "@/lib/status";
import type { SustainabilityClaim } from "@/lib/types";

const ANY = "any";

const claimStatuses = Object.keys(claimStatusPresentation) as ClaimStatus[];

const columns: DataTableColumn<SustainabilityClaim>[] = [
  {
    key: "activity",
    header: "Activity",
    render: (claim) => (
      <span className="block">
        <Link
          href={`/claims/${claim.id}`}
          className="font-medium underline-offset-4 hover:underline"
        >
          {activityTypeLabels[claim.activityType]}
        </Link>
        <span className="block text-caption text-neutral-600">
          {claim.facilityName}
        </span>
      </span>
    ),
  },
  {
    key: "period",
    header: "Period",
    render: (claim) => (
      <span className="text-helper">
        <TimestampDisplay value={claim.periodStart} dateOnly /> —{" "}
        <TimestampDisplay value={claim.periodEnd} dateOnly />
      </span>
    ),
  },
  {
    key: "submitted",
    header: "Submitted",
    hideOnCard: true,
    render: (claim) => <TimestampDisplay value={claim.submittedAt} dateOnly />,
  },
  {
    key: "ceiling",
    header: "Ceiling",
    alignEnd: true,
    render: (claim) => <CreditAmountDisplay amount={claim.ceiling.ceiling} />,
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

export function ClaimsTable({ claims }: { claims: SustainabilityClaim[] }) {
  const [status, setStatus] = useState<string>(ANY);
  const [activity, setActivity] = useState<string>(ANY);

  const filtered = useMemo(
    () =>
      claims.filter(
        (claim) =>
          (status === ANY || claim.status === status) &&
          (activity === ANY || claim.activityType === activity),
      ),
    [claims, status, activity],
  );

  const activeCount = [status, activity].filter(
    (value) => value !== ANY,
  ).length;

  const clear = () => {
    setStatus(ANY);
    setActivity(ANY);
  };

  return (
    <div className="space-y-4">
      <FilterBar activeCount={activeCount} onClear={clear}>
        <Select value={status} onValueChange={setStatus}>
          <SelectTrigger className="w-48">
            <SelectValue placeholder="Any status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ANY}>Any status</SelectItem>
            {claimStatuses.map((name) => (
              <SelectItem key={name} value={name}>
                {claimStatusPresentation[name].label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select value={activity} onValueChange={setActivity}>
          <SelectTrigger className="w-56">
            <SelectValue placeholder="Any activity" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ANY}>Any activity</SelectItem>
            {Object.entries(activityTypeLabels).map(([value, label]) => (
              <SelectItem key={value} value={value}>
                {label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </FilterBar>

      <DataTable
        columns={columns}
        rows={filtered}
        rowKey={(claim) => claim.id}
        caption="Sustainability claims filed by your organization"
        emptyState={
          <EmptyState
            icon={FileCheck2}
            title={
              activeCount > 0
                ? "No claims match those filters"
                : "No claims yet"
            }
            description="A sustainability claim converts a facility's verified activity into carbon credits, subject to a computed ceiling and verifier approval."
            action={
              <Button asChild>
                <Link href="/claims/new">Submit a claim</Link>
              </Button>
            }
          />
        }
      />
    </div>
  );
}
