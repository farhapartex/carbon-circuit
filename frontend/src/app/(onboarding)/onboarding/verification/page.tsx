import type { Metadata } from "next";
import { OnboardingStepper } from "@/components/features/onboarding/OnboardingStepper";
import { VerificationOutcome } from "@/components/features/onboarding/VerificationOutcome";
import { PageHeader } from "@/components/shared/PageHeader";

export const metadata: Metadata = { title: "Registry verification" };

export default function OnboardingVerificationPage() {
  return (
    <>
      <OnboardingStepper current="verification" />
      <PageHeader
        title="Registry verification"
        description="We check your registration number against the business register. This is what stops someone registering a fictional company with an enormous declared capacity and minting credits against it."
      />
      <VerificationOutcome />
    </>
  );
}
