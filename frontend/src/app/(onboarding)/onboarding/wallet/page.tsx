import type { Metadata } from "next";
import { OnboardingStepper } from "@/components/features/onboarding/OnboardingStepper";
import { TreasuryDesignation } from "@/components/features/onboarding/TreasuryDesignation";
import { PageHeader } from "@/components/shared/PageHeader";

export const metadata: Metadata = { title: "Connect your wallet" };

export default function OnboardingWalletPage() {
  return (
    <>
      <OnboardingStepper current="wallet" />
      <PageHeader
        title="Connect your wallet"
        description="The last step. This address becomes your organization's Treasury Address, where every credit you earn is delivered and from where every sale and retirement happens."
      />
      <TreasuryDesignation />
    </>
  );
}
