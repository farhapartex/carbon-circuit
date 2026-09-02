import type { Metadata } from "next";
import { OnboardingStepper } from "@/components/features/onboarding/OnboardingStepper";
import { PlanSelector } from "@/components/features/onboarding/PlanSelector";
import { PageHeader } from "@/components/shared/PageHeader";
import { fetchPlans } from "@/lib/api/plans";
import { auth0 } from "@/lib/auth0";
import { tenancyOf } from "@/lib/session";

export const metadata: Metadata = { title: "Choose a plan" };

export default async function OnboardingPlanPage() {
  const tenancy = tenancyOf(await auth0.getSession());
  const plans = await fetchPlans(tenancy?.organizationType ?? undefined);

  return (
    <>
      <OnboardingStepper current="plan" />
      <PageHeader
        title="Choose a plan"
        description="Showing only the plans your organization type can use. Organizations that just buy and retire credits pay nothing, because a marketplace with no buyers is worthless to the sellers who do pay."
      />
      <PlanSelector plans={plans} />
    </>
  );
}
