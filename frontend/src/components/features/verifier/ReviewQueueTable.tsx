"use client";

import Link from "next/link";
import { EmptyState } from "@/components/shared/EmptyState";
import { PriorityBadge } from "@/components/shared/StatusBadges";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Button } from "@/components/ui/button";
import type { ClaimRecord } from "@/lib/api/claims";
import { activityTypeLabels } from "@/lib/labels";

const decimalFormat = new Intl.NumberFormat("en-US", {
  maximumFractionDigits: 2,
});

const amount = (value: string) => decimalFormat.format(Number(value));

export function ReviewQueueTable({ claims }: { claims: ClaimRecord[] }) {
  if (claims.length === 0) {
    return (
      <EmptyState
        title="Nothing is waiting for review"
        description="Claims appear here once they have been submitted and have come back from AI review."
      />
    );
  }

  return (
    <div className="overflow-x-auto rounded-lg border border-neutral-200 bg-white">
      <table className="w-full min-w-4xl border-collapse text-left">
        <thead className="border-b border-neutral-200">
          <tr>
            <th scope="col" className="px-4 py-3 text-caption font-medium">
              Priority
            </th>
            <th scope="col" className="px-4 py-3 text-caption font-medium">
              Facility
            </th>
            <th scope="col" className="px-4 py-3 text-caption font-medium">
              Activity
            </th>
            <th
              scope="col"
              className="px-4 py-3 text-right text-caption font-medium"
            >
              Requested
            </th>
            <th
              scope="col"
              className="px-4 py-3 text-right text-caption font-medium"
            >
              Ceiling
            </th>
            <th scope="col" className="px-4 py-3 text-caption font-medium">
              Submitted
            </th>
            <th scope="col" className="px-4 py-3" />
          </tr>
        </thead>
        <tbody className="divide-y divide-neutral-200">
          {claims.map((claim) => (
            <tr key={claim.id}>
              <td className="px-4 py-3">
                <PriorityBadge priority={claim.priority} />
                {claim.requiresDualApproval ? (
                  <span className="mt-1 block text-caption text-neutral-600">
                    two approvals
                  </span>
                ) : null}
              </td>
              <td className="px-4 py-3">
                <span className="block font-medium">{claim.facilityName}</span>
                <span className="block text-caption text-neutral-600">
                  vintage {claim.vintageYear}
                </span>
              </td>
              <td className="px-4 py-3 text-caption">
                {activityTypeLabels[claim.activityType]}
              </td>
              <td className="px-4 py-3 text-right tabular-nums">
                {amount(claim.requestedAmount)}
              </td>
              <td className="px-4 py-3 text-right tabular-nums">
                {amount(claim.computedCeiling)}
                {Number(claim.requestedAmount) >
                Number(claim.computedCeiling) ? (
                  <span className="mt-1 block text-caption text-warning-700">
                    over ceiling
                  </span>
                ) : null}
              </td>
              <td className="px-4 py-3 text-caption text-neutral-600">
                <TimestampDisplay value={claim.createdAt} dateOnly />
              </td>
              <td className="px-4 py-3 text-right">
                <Button asChild variant="outline" size="sm">
                  <Link href={`/verifier/queue/${claim.id}`}>Review</Link>
                </Button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
