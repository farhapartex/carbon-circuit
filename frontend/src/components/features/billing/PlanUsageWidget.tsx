import { StatusPill } from "@/components/shared/StatusPill";
import type { PlanUsage, UsageDimension } from "@/lib/types";

const percentageOf = (dimension: UsageDimension) =>
  dimension.limit === null
    ? 0
    : Math.min(100, Math.round((dimension.used / dimension.limit) * 100));

const toneFor = (dimension: UsageDimension) => {
  const percentage = percentageOf(dimension);
  if (percentage >= 100)
    return { label: "Exhausted", variant: "danger" } as const;
  if (percentage >= 80)
    return { label: "Near limit", variant: "warning" } as const;
  return null;
};

const numberFormat = new Intl.NumberFormat("en-US");

export function PlanUsageWidget({ usage }: { usage: PlanUsage }) {
  return (
    <dl className="space-y-4">
      {usage.dimensions.map((dimension) => {
        const tone = toneFor(dimension);
        return (
          <div key={dimension.key} className="space-y-1.5">
            <div className="flex flex-wrap items-baseline justify-between gap-2">
              <dt className="flex items-center gap-2 text-caption font-medium">
                {dimension.label}
                {tone ? <StatusPill presentation={tone} /> : null}
              </dt>
              <dd className="text-caption text-neutral-600 tabular-nums">
                {numberFormat.format(dimension.used)}
                {dimension.limit === null
                  ? " used"
                  : ` of ${numberFormat.format(dimension.limit)}`}
              </dd>
            </div>
            {dimension.limit === null ? null : (
              <progress
                data-slot="score-meter"
                value={dimension.used}
                max={dimension.limit}
                aria-label={`${dimension.label}: ${dimension.used} of ${dimension.limit}`}
              />
            )}
            <p className="text-caption text-neutral-600">
              {dimension.blocksOnExhaustion
                ? "Reaching this limit blocks new records until you upgrade."
                : dimension.overageRateUsd
                  ? `Beyond the included amount you are billed $${dimension.overageRateUsd} per additional claim.`
                  : "Crossing this starts a conversation rather than blocking you."}
            </p>
          </div>
        );
      })}
    </dl>
  );
}
