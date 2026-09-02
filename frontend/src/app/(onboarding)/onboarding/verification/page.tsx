import type { Metadata } from "next";
import { OnboardingStepper } from "@/components/features/onboarding/OnboardingStepper";
import { VerificationOutcome } from "@/components/features/onboarding/VerificationOutcome";
import { PageHeader } from "@/components/shared/PageHeader";
import { fetchCurrentOrganization } from "@/lib/api/organization";
import { auth0 } from "@/lib/auth0";

export const metadata: Metadata = { title: "Registry verification" };

export default async function OnboardingVerificationPage() {
  const { token } = await auth0.getAccessToken();
  const organization = await fetchCurrentOrganization(token);

  return (
    <>
      <OnboardingStepper current="verification" />
      <PageHeader
        title="Registry verification"
        description="We check your registration number against the business register. This is what stops someone registering a fictional company with an enormous declared capacity and minting credits against it."
      />
      <VerificationOutcome
        declaredName={organization.name}
        countryCode={organization.countryOfIncorporation}
        registrationNumber={organization.businessRegistrationNumber}
        status={organization.outcome.status}
        rejection={organization.outcome.rejection}
        nameSimilarity={organization.outcome.nameSimilarity}
        registeredLegalName={organization.outcome.registeredLegalName}
      />
    </>
  );
}
