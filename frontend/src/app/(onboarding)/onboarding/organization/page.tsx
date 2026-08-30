import type { Metadata } from "next";
import { OnboardingStepper } from "@/components/features/onboarding/OnboardingStepper";
import { OrganizationForm } from "@/components/features/onboarding/OrganizationForm";
import { PageHeader } from "@/components/shared/PageHeader";

export const metadata: Metadata = { title: "Create your organization" };

export default function OnboardingOrganizationPage() {
  return (
    <>
      <OnboardingStepper current="organization" />
      <PageHeader
        title="Create your organization"
        description="Everything on CarbonCircuit belongs to an organization rather than to a person. You will be its Owner, which is the role that can later change where your credits are held."
      />
      <OrganizationForm />
    </>
  );
}
