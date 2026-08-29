import type { Metadata } from "next";
import {
  Calculator,
  Fingerprint,
  Flame,
  Quote,
  ScrollText,
  Store,
} from "lucide-react";
import { CTABanner } from "@/components/features/marketing/CTABanner";
import { FeatureGrid } from "@/components/features/marketing/FeatureGrid";
import { HeroSection } from "@/components/features/marketing/HeroSection";
import { HowItWorksSteps } from "@/components/features/marketing/HowItWorksSteps";
import { TrustBand } from "@/components/features/marketing/TrustBand";

export const metadata: Metadata = {
  title: "Carbon credits",
  description:
    "Independently reviewed carbon credits that permanently name the facility that earned them.",
};

const features = [
  {
    icon: Quote,
    title: "AI review with citations, human decisions",
    description:
      "The agent extracts each figure with the document and page it came from. A verifier checks the cited page in seconds. The AI never approves a claim and never computes the credit amount.",
  },
  {
    icon: Calculator,
    title: "A published formula, not a negotiation",
    description:
      "Credits come from a defined formula per activity type, drawing multipliers from a maintained reference table. The version used is pinned to the claim forever, so a recomputation years later gives the same answer.",
  },
  {
    icon: Fingerprint,
    title: "Credit Class attribution",
    description:
      "Originating facility, vintage year, and activity type travel with the credit through every transfer and are still attached at retirement. Different classes are never pooled or averaged.",
  },
  {
    icon: Store,
    title: "A marketplace that cannot oversell",
    description:
      "Listing moves credits into escrow, so they leave the seller's spendable balance the moment the listing goes live. Overselling is structurally impossible rather than merely checked for.",
  },
  {
    icon: Flame,
    title: "Retirement that means something",
    description:
      "Retiring burns the credit and records who retired it, what it represented, and why. It can never be resold, retraded, or retired again.",
  },
  {
    icon: ScrollText,
    title: "An audit trail built for seven years",
    description:
      "Every AI review keeps its full trace: node states, model identifier, program version, every tool call, and the reference table version. An auditor asking why a claim was recommended gets the whole answer.",
  },
];

const steps = [
  {
    title: "A facility submits a claim",
    description:
      "Declared figures, the claim period, supporting evidence, and a recorded exclusivity attestation tied to the submitting user.",
  },
  {
    title: "The ceiling is computed",
    description:
      "A claim can never issue more than its ceiling. A facility with no registry match is discounted to half of what its self-declared capacity would support.",
  },
  {
    title: "AI review, then human decision",
    description:
      "Figures are extracted with citations and cross-checked against reference ranges and prior evidence. A verifier approves, rejects, or asks for more. Above 5,000 tCO2e it takes two.",
  },
  {
    title: "Credits are issued and tradeable",
    description:
      "Minted to the organization's Treasury Address against an authorization the contract verifies itself, then listed, bought, and eventually retired.",
  },
];

export default function CarbonCreditsPage() {
  return (
    <>
      <HeroSection
        eyebrow="For sustainability teams and credit buyers"
        title="Credits you can name a facility for"
        description="A buyer's whole reason for paying a premium is being able to point at a specific verified practice in their own ESG report. A credit pooled into an anonymous balance cannot support that claim, so CarbonCircuit never pools them."
        primaryAction={{ label: "Browse listings", href: "/marketplace" }}
        secondaryAction={{
          label: "Audit the retirement log",
          href: "/marketplace/retirements",
        }}
      />
      <FeatureGrid
        heading="Why these credits are different"
        description="Carbon markets have been damaged by credits that could not be traced or were counted twice. Every decision below exists to make that harder here."
        features={features}
      />
      <HowItWorksSteps heading="From practice to retirement" steps={steps} />
      <TrustBand />
      <CTABanner
        title="Buying and retiring is free"
        description="Organizations that only purchase and retire credits pay no subscription. Sellers pay a transaction fee on completed trades."
        primaryAction={{ label: "Start buying", href: "/signup" }}
        secondaryAction={{ label: "Sell your credits", href: "/pricing" }}
      />
    </>
  );
}
