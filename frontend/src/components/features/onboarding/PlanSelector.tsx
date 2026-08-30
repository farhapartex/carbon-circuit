"use client";

import { Check } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { StatusPill } from "@/components/shared/StatusPill";
import { Button } from "@/components/ui/button";
import type { OrganizationType, Plan, PlanTier } from "@/lib/types";
import { cn } from "@/lib/utils";
import { useFormDraftStore } from "@/stores/form-drafts";

const isFree = (plan: Plan) => plan.monthlyPriceUsd === "0";

const cheapestOf = (plans: Plan[]) =>
  [...plans].sort(
    (left, right) =>
      Number(left.monthlyPriceUsd ?? 0) - Number(right.monthlyPriceUsd ?? 0),
  )[0];

const summaryOf = (plan: Plan) =>
  [
    plan.batchesPerMonth?.included
      ? `${plan.batchesPerMonth.included.toLocaleString()} batches / month`
      : plan.batchesPerMonth
        ? "Unlimited batches"
        : null,
    plan.aiReviewedClaimsPerMonth?.included
      ? `${plan.aiReviewedClaimsPerMonth.included} AI-reviewed claims / month`
      : null,
    plan.apiRateLimitPerMinute ? "Portal and API submission" : "Portal entry",
    plan.marketplaceFeeBasisPoints === null
      ? "No marketplace fee"
      : `${(plan.marketplaceFeeBasisPoints / 100).toFixed(1)}% seller fee`,
  ].filter((line): line is string => line !== null);

export function PlanSelector({ plans }: { plans: Plan[] }) {
  const router = useRouter();
  const saveDraft = useFormDraftStore((state) => state.saveDraft);
  const draft = useFormDraftStore((state) => state.drafts.organization);
  const organizationType = (draft?.values.type ??
    "manufacturer") as OrganizationType;

  const eligible = plans.filter((plan) =>
    plan.allowedOrganizationTypes.includes(organizationType),
  );
  const freePlan = eligible.find(isFree);
  const defaultTier = (freePlan ?? cheapestOf(eligible))?.tier;

  const [selected, setSelected] = useState<PlanTier | undefined>(defaultTier);
  const selectedPlan = eligible.find((plan) => plan.tier === selected);

  const confirm = () => {
    if (!selectedPlan) return;
    saveDraft("organization", {
      step: 3,
      values: { ...(draft?.values ?? {}), planTier: selectedPlan.tier },
      evidenceIds: [],
    });
    router.push("/onboarding/wallet");
  };

  return (
    <div className="space-y-6">
      <ul className="grid gap-4 sm:grid-cols-2">
        {eligible.map((plan) => {
          const active = plan.tier === selected;
          return (
            <li key={plan.tier}>
              <button
                type="button"
                onClick={() => setSelected(plan.tier)}
                aria-pressed={active}
                className={cn(
                  "flex h-full w-full flex-col gap-3 rounded-lg border p-6 text-left transition-colors",
                  active
                    ? "border-primary-600 bg-primary-50"
                    : "border-neutral-200 bg-white hover:border-neutral-400",
                )}
              >
                <span className="flex items-start justify-between gap-3">
                  <span className="font-medium">{plan.name}</span>
                  {isFree(plan) ? (
                    <StatusPill
                      presentation={{ label: "Free", variant: "success" }}
                      showDot={false}
                    />
                  ) : null}
                </span>

                <span className="text-section-heading tabular-nums">
                  {isFree(plan) ? "Free" : `$${plan.monthlyPriceUsd}`}
                  {isFree(plan) ? null : (
                    <span className="text-caption font-normal text-neutral-600">
                      {" "}
                      / month
                    </span>
                  )}
                </span>

                <span className="text-caption text-pretty text-neutral-600">
                  {plan.audience}
                </span>

                <span className="mt-auto space-y-1">
                  {summaryOf(plan).map((line) => (
                    <span
                      key={line}
                      className="flex items-start gap-1.5 text-caption text-neutral-600"
                    >
                      <Check
                        className="mt-0.5 size-3.5 shrink-0 text-primary-600"
                        aria-hidden
                      />
                      {line}
                    </span>
                  ))}
                </span>
              </button>
            </li>
          );
        })}
      </ul>

      <div className="flex flex-wrap items-center gap-4">
        <Button size="lg" onClick={confirm} disabled={!selectedPlan}>
          {selectedPlan && isFree(selectedPlan)
            ? "Continue with the free plan"
            : "Continue to payment"}
        </Button>
        {selectedPlan && !isFree(selectedPlan) ? (
          <p className="text-caption text-neutral-600">
            You will be taken to Stripe. Nothing is charged until you confirm
            there.
          </p>
        ) : null}
      </div>
    </div>
  );
}
