import { Check, Minus } from "lucide-react";
import type { ReactNode } from "react";
import type { Plan, PlanLimit } from "@/lib/types";
import { cn } from "@/lib/utils";

const numberFormat = new Intl.NumberFormat("en-US");

const describeLimit = (limit: PlanLimit | null): ReactNode => {
  if (!limit) return <Minus className="size-4 text-neutral-400" aria-hidden />;
  if (limit.included !== null) return numberFormat.format(limit.included);
  if (limit.fairUseCeiling !== null) {
    return (
      <span>
        Unlimited
        <span className="block text-caption text-neutral-600">
          {numberFormat.format(limit.fairUseCeiling)} fair-use ceiling
        </span>
      </span>
    );
  }
  return <Check className="size-4 text-success-600" aria-hidden />;
};

const describeFee = (basisPoints: number | null) =>
  basisPoints === null ? "No fee" : `${(basisPoints / 100).toFixed(1)}%`;

const describeRate = (perMinute: number | null) =>
  perMinute === null ? (
    <Minus className="size-4 text-neutral-400" aria-hidden />
  ) : (
    `${numberFormat.format(perMinute)} req/min`
  );

const rows: { label: string; render: (plan: Plan) => ReactNode }[] = [
  { label: "Who it's for", render: (plan) => plan.audience },
  {
    label: "Batches per month",
    render: (plan) => describeLimit(plan.batchesPerMonth),
  },
  {
    label: "Checkpoints per month",
    render: (plan) => describeLimit(plan.checkpointsPerMonth),
  },
  { label: "Facilities", render: (plan) => describeLimit(plan.facilities) },
  { label: "Users", render: (plan) => describeLimit(plan.users) },
  {
    label: "AI-reviewed claims per month",
    render: (plan) => describeLimit(plan.aiReviewedClaimsPerMonth),
  },
  {
    label: "Beyond the claim quota",
    render: (plan) =>
      plan.aiReviewedClaimsPerMonth?.overageRateUsd
        ? `$${plan.aiReviewedClaimsPerMonth.overageRateUsd} per additional claim`
        : plan.aiReviewedClaimsPerMonth
          ? "Upgrade required"
          : "—",
  },
  {
    label: "Evidence storage",
    render: (plan) =>
      plan.evidenceStorageGb === null ? "—" : `${plan.evidenceStorageGb} GB`,
  },
  {
    label: "Data submission",
    render: (plan) =>
      plan.apiRateLimitPerMinute === null
        ? plan.batchesPerMonth
          ? "Portal only"
          : "—"
        : "Portal and API",
  },
  {
    label: "Portal rate limit",
    render: (plan) => describeRate(plan.portalRateLimitPerMinute),
  },
  {
    label: "API rate limit",
    render: (plan) => describeRate(plan.apiRateLimitPerMinute),
  },
  {
    label: "API keys",
    render: (plan) =>
      plan.apiKeyLimit === null ? (
        <Minus className="size-4 text-neutral-400" aria-hidden />
      ) : (
        `${plan.apiKeyLimit} active`
      ),
  },
  {
    label: "Marketplace fee, charged to seller",
    render: (plan) => describeFee(plan.marketplaceFeeBasisPoints),
  },
  { label: "Review turnaround", render: (plan) => plan.reviewTurnaround },
  { label: "Support", render: (plan) => plan.supportLevel },
];

const priceOf = (plan: Plan) =>
  plan.monthlyPriceUsd === "0" ? "Free" : `$${plan.monthlyPriceUsd}`;

type PlanComparisonTableProps = {
  plans: Plan[];
  action?: ((plan: Plan) => ReactNode) | undefined;
  highlightTier?: Plan["tier"] | undefined;
};

export function PlanComparisonTable({
  plans,
  action,
  highlightTier,
}: PlanComparisonTableProps) {
  return (
    <div data-slot="plan-comparison">
      <div className="hidden overflow-x-auto lg:block">
        <table className="w-full border-collapse text-caption">
          <caption className="sr-only">Plan comparison</caption>
          <thead>
            <tr>
              <th scope="col" className="w-56 p-4 text-left align-bottom" />
              {plans.map((plan) => (
                <th
                  key={plan.tier}
                  scope="col"
                  className={cn(
                    "space-y-2 border-b border-neutral-200 p-4 text-left align-bottom",
                    plan.tier === highlightTier && "bg-primary-50",
                  )}
                >
                  <span className="block text-body font-medium">
                    {plan.name}
                  </span>
                  <span className="block text-section-heading tabular-nums">
                    {priceOf(plan)}
                    {plan.monthlyPriceUsd !== "0" ? (
                      <span className="text-caption font-normal text-neutral-600">
                        {" "}
                        / month
                      </span>
                    ) : null}
                  </span>
                  {plan.priceNote ? (
                    <span className="block font-normal text-neutral-600">
                      {plan.priceNote}
                    </span>
                  ) : null}
                  {action ? (
                    <span className="block">{action(plan)}</span>
                  ) : null}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.label} className="border-b border-neutral-200">
                <th
                  scope="row"
                  className="p-4 text-left font-medium text-neutral-600"
                >
                  {row.label}
                </th>
                {plans.map((plan) => (
                  <td
                    key={plan.tier}
                    className={cn(
                      "p-4 align-top",
                      plan.tier === highlightTier && "bg-primary-50",
                    )}
                  >
                    {row.render(plan)}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <ul className="grid gap-4 sm:grid-cols-2 lg:hidden">
        {plans.map((plan) => (
          <li
            key={plan.tier}
            className={cn(
              "space-y-4 rounded-lg border border-neutral-200 bg-white p-6",
              plan.tier === highlightTier && "border-primary-600",
            )}
          >
            <div className="space-y-1">
              <p className="font-medium">{plan.name}</p>
              <p className="text-section-heading tabular-nums">
                {priceOf(plan)}
              </p>
              <p className="text-caption text-neutral-600">{plan.audience}</p>
            </div>
            {action ? action(plan) : null}
            <dl className="space-y-2">
              {rows.slice(1).map((row) => (
                <div
                  key={row.label}
                  className="flex items-baseline justify-between gap-4"
                >
                  <dt className="text-caption text-neutral-600">{row.label}</dt>
                  <dd className="text-right text-caption">
                    {row.render(plan)}
                  </dd>
                </div>
              ))}
            </dl>
          </li>
        ))}
      </ul>
    </div>
  );
}
