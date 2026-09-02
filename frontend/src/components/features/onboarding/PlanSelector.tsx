"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { PlanComparisonTable } from "@/components/features/billing/PlanComparisonTable";
import { Button } from "@/components/ui/button";
import type { Plan, PlanTier } from "@/lib/types";
import { submitSubscription } from "@/lib/actions/subscription";
import { useFormDraftStore } from "@/stores/form-drafts";

const isFree = (plan: Plan) => plan.monthlyPriceUsd === "0";

const cheapestOf = (plans: Plan[]) =>
  [...plans].sort(
    (left, right) =>
      Number(left.monthlyPriceUsd ?? 0) - Number(right.monthlyPriceUsd ?? 0),
  )[0];

export function PlanSelector({ plans }: { plans: Plan[] }) {
  const router = useRouter();
  const saveDraft = useFormDraftStore((state) => state.saveDraft);
  const draft = useFormDraftStore((state) => state.drafts.organization);

  const preferred = (plans.find(isFree) ?? cheapestOf(plans))?.tier;
  const [selected, setSelected] = useState<PlanTier | undefined>(preferred);
  const [pending, startTransition] = useTransition();
  const [failure, setFailure] = useState<string | null>(null);
  const [idempotencyKey] = useState(() => crypto.randomUUID());
  const selectedPlan = plans.find((plan) => plan.tier === selected);

  const confirm = () => {
    if (!selectedPlan) return;

    setFailure(null);
    saveDraft("organization", {
      step: 3,
      values: { ...(draft?.values ?? {}), planTier: selectedPlan.tier },
      evidenceIds: [],
    });

    startTransition(async () => {
      const result = await submitSubscription(
        selectedPlan.tier,
        idempotencyKey,
      );

      if (!result.ok) {
        setFailure(subscriptionFailure(result.code));
        return;
      }

      router.push("/onboarding/wallet");
    });
  };

  return (
    <div className="space-y-6">
      {failure ? (
        <div
          role="alert"
          className="rounded-md border border-danger-600 bg-danger-50 px-4 py-3"
        >
          <p className="font-700 text-body text-danger-700">
            We could not start your subscription
          </p>
          <p className="text-helper text-danger-700">{failure}</p>
        </div>
      ) : null}

      <PlanComparisonTable
        plans={plans}
        highlightTier={selected}
        action={(plan) => (
          <Button
            size="sm"
            variant={plan.tier === selected ? "default" : "outline"}
            onClick={() => setSelected(plan.tier)}
            aria-pressed={plan.tier === selected}
          >
            {plan.tier === selected ? "Selected" : "Select"}
          </Button>
        )}
      />

      <div className="flex flex-wrap items-center gap-4">
        <Button size="lg" onClick={confirm} disabled={!selectedPlan || pending}>
          {pending
            ? "Starting your subscription…"
            : selectedPlan && isFree(selectedPlan)
              ? "Continue with the free plan"
              : `Continue with ${selectedPlan?.name ?? "a plan"}`}
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

function subscriptionFailure(code: string): string {
  switch (code) {
    case "CONFLICT":
      return "This organization already has a subscription.";
    case "FORBIDDEN":
      return "Only an owner or admin can choose the plan for an organization.";
    case "REQUEST_IN_PROGRESS":
      return "This selection is already being processed. Give it a moment.";
    default:
      return "Something went wrong on our side. Try again shortly.";
  }
}
