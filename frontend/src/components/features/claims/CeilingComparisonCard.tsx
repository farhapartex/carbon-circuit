import { TriangleAlert } from "lucide-react";
import { CreditAmountDisplay } from "@/components/shared/CreditAmountDisplay";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { compareCreditAmounts } from "@/lib/decimal";
import type { SustainabilityClaim } from "@/lib/types";

export function CeilingComparisonCard({
  claim,
}: {
  claim: SustainabilityClaim;
}) {
  const exceedsCeiling =
    compareCreditAmounts(claim.requestedAmount, claim.ceiling.ceiling) > 0;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Requested against ceiling</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <div>
            <p className="text-caption text-neutral-600">Requested</p>
            <p className="text-lg font-medium">
              <CreditAmountDisplay amount={claim.requestedAmount} />
            </p>
          </div>
          <div>
            <p className="text-caption text-neutral-600">Computed ceiling</p>
            <p className="text-lg font-medium">
              <CreditAmountDisplay amount={claim.ceiling.ceiling} />
            </p>
          </div>
        </div>

        {exceedsCeiling ? (
          <div className="rounded-md border border-warning-600 bg-warning-50 px-4 py-3">
            <p className="flex items-center gap-2 font-medium text-warning-700">
              <TriangleAlert className="size-4 shrink-0" aria-hidden />
              This claim requests more than its ceiling
            </p>
            <p className="mt-1 text-caption text-pretty text-warning-700">
              A claim can never issue more than its ceiling regardless of what
              was requested. A verifier cannot approve above it.
            </p>
          </div>
        ) : null}

        <dl className="space-y-3 border-t border-neutral-200 pt-4">
          <div className="flex flex-wrap items-baseline justify-between gap-4">
            <dt className="text-caption text-neutral-600">Discount factor</dt>
            <dd className="font-medium tabular-nums">
              {claim.ceiling.discountFactor}
            </dd>
          </div>
          <p className="text-caption text-pretty text-neutral-600">
            {claim.ceiling.discountReason}
          </p>
          <div className="flex flex-wrap items-baseline justify-between gap-4">
            <dt className="text-caption text-neutral-600">Reference factor</dt>
            <dd className="font-medium">
              {claim.ceiling.referenceFactorValue}{" "}
              {claim.ceiling.referenceFactorUnit}
            </dd>
          </div>
          <div className="flex flex-wrap items-baseline justify-between gap-4">
            <dt className="text-caption text-neutral-600">
              Reference table version
            </dt>
            <dd className="font-mono text-helper">
              {claim.ceiling.referenceTableVersion}
            </dd>
          </div>
        </dl>

        <p className="text-caption text-pretty text-neutral-600">
          The reference table version is pinned to this claim permanently, so
          recomputing it years from now produces the same answer.
        </p>
      </CardContent>
    </Card>
  );
}
