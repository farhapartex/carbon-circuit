import type { Metadata } from "next";
import Link from "next/link";
import { BuyerPlanCallout } from "@/components/features/billing/BuyerPlanCallout";
import { PlanComparisonTable } from "@/components/features/billing/PlanComparisonTable";
import { Button } from "@/components/ui/button";
import { fetchPlans } from "@/lib/api/plans";

export const metadata: Metadata = {
  title: "Pricing",
  description:
    "Plans for facilities, logistics partners, and enterprises. Buying and retiring credits is free.",
};

export default async function PricingPage() {
  const plans = await fetchPlans();

  return (
    <div className="mx-auto max-w-6xl space-y-12 px-6 py-16">
      <header className="max-w-2xl space-y-3">
        <h1 className="text-page-title">Pricing</h1>
        <p className="text-body text-pretty text-neutral-600">
          Organizations that produce data pay a subscription. Organizations that
          only buy and retire credits pay nothing. Every AI-assisted claim
          review is included in the subscription rather than billed as a
          separate line item.
        </p>
      </header>

      <BuyerPlanCallout />

      <PlanComparisonTable
        plans={plans}
        highlightTier="growth"
        action={(plan) => (
          <Button
            asChild
            size="sm"
            variant={plan.tier === "growth" ? "default" : "outline"}
          >
            <Link href="/signup">
              {plan.tier === "enterprise" ? "Contact sales" : "Get started"}
            </Link>
          </Button>
        )}
      />

      <section className="space-y-3 rounded-lg border border-neutral-200 bg-white p-6">
        <h2 className="font-medium">Two things worth knowing</h2>
        <p className="text-caption text-pretty text-neutral-600">
          A claim sent back by a verifier for more information does not count
          against your monthly quota when you resubmit it. Only genuinely new
          submissions do, so you are never penalised for a reviewer&apos;s
          request.
        </p>
        <p className="text-caption text-pretty text-neutral-600">
          Fair-use ceilings on Enterprise exist so that &ldquo;unlimited&rdquo;
          has a defined operational meaning. Crossing one does not block your
          organization. It starts a conversation, because silently throttling a
          paying customer mid-shipment is worse than a phone call.
        </p>
      </section>
    </div>
  );
}
