import type { Metadata } from "next";
import { PlanComparisonTable } from "@/components/features/billing/PlanComparisonTable";
import { Button } from "@/components/ui/button";
import {
  getCurrentOrganization,
  getSubscription,
  listPlans,
} from "@/lib/fixtures";

export const metadata: Metadata = { title: "Plans" };

export default async function SettingsBillingPlansPage() {
  const plans = await listPlans();
  const subscription = await getSubscription();
  const organization = await getCurrentOrganization();

  const eligiblePlans = plans.filter((plan) =>
    plan.allowedOrganizationTypes.includes(organization.type),
  );

  return (
    <>
      <p className="max-w-2xl text-caption text-pretty text-neutral-600">
        Showing the plans available to a {organization.type.replace("_", " ")}{" "}
        organization. Changing plan takes effect immediately, and any difference
        is prorated against your current billing period.
      </p>

      <PlanComparisonTable
        plans={eligiblePlans}
        highlightTier={subscription.planTier}
        action={(plan) =>
          plan.tier === subscription.planTier ? (
            <Button size="sm" variant="outline" disabled>
              Current plan
            </Button>
          ) : (
            <Button size="sm">
              {plan.tier === "enterprise" ? "Contact sales" : "Switch to this"}
            </Button>
          )
        }
      />
    </>
  );
}
