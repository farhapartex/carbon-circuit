import type { Metadata } from "next";
import { OnboardingStepper } from "@/components/features/onboarding/OnboardingStepper";
import { PlanSelector } from "@/components/features/onboarding/PlanSelector";
import { PageHeader } from "@/components/shared/PageHeader";
import { listPlans } from "@/lib/fixtures";

export const metadata: Metadata = { title: "Choose a plan" };

export default async function OnboardingPlanPage() {
  const plans = await listPlans();

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
