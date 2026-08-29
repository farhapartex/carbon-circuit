import { FileCheck2, Layers, Link2, Lock, ScanLine, Users } from "lucide-react";
import { CTABanner } from "@/components/features/marketing/CTABanner";
import { FeatureGrid } from "@/components/features/marketing/FeatureGrid";
import { HeroSection } from "@/components/features/marketing/HeroSection";
import { PersonaSplitSection } from "@/components/features/marketing/PersonaSplitSection";
import { StatsBand } from "@/components/features/marketing/StatsBand";

const features = [
  {
    icon: ScanLine,
    title: "Checkpoints that stay honest",
    description:
      "A logged checkpoint is never edited or deleted. A correction supersedes the original and both stay visible, because silently rewriting a supply chain record is the failure this product exists to prevent.",
  },
  {
    icon: Link2,
    title: "Anchored to a public ledger",
    description:
      "Every batch and checkpoint across the platform is covered by one Merkle root per ten-minute epoch. Historical roots are kept forever, so a proof written today still verifies in a decade.",
  },
  {
    icon: FileCheck2,
    title: "Reviewed by a person, assisted by AI",
    description:
      "An AI agent reads the evidence and cites the page it drew each figure from. A human verifier makes every approval, and the credit amount comes from a deterministic formula the model cannot reach.",
  },
  {
    icon: Layers,
    title: "Credits that keep their identity",
    description:
      "Facility, vintage, and activity type are properties of the token itself, not bookkeeping alongside it. They cannot drift apart, because there is no second ledger to drift from.",
  },
  {
    icon: Lock,
    title: "Retirement is final",
    description:
      "A retired credit can never be resold, retraded, or retired again by anyone. That is what stops the same tonne being claimed twice on this platform.",
  },
  {
    icon: Users,
    title: "Built for competitors sharing one platform",
    description:
      "Tenant isolation is enforced by the database, not only by application code. Public identifiers are random, so nobody can infer another company's volume from a URL.",
  },
];

const stats = [
  { value: "5", label: "Provenance Score components, each explained" },
  {
    value: "3",
    label: "Sustainability activity types with published formulas",
  },
  { value: "18", label: "Grid regions in the emission factor table" },
  { value: "10 min", label: "Epoch between on-chain anchors, platform-wide" },
];

export default function HomePage() {
  return (
    <>
      <HeroSection
        eyebrow="Provenance and carbon credits"
        title="Prove where it came from. Prove what it saved."
        description="CarbonCircuit records the verifiable journey of a batch through every facility that touches it, and turns independently reviewed sustainability practice into carbon credits that never lose the name of the facility that earned them."
        primaryAction={{ label: "Get started", href: "/signup" }}
        secondaryAction={{
          label: "Browse the marketplace",
          href: "/marketplace",
        }}
      />
      <PersonaSplitSection />
      <StatsBand stats={stats} />
      <FeatureGrid
        heading="Two systems, one chain of evidence"
        description="Provenance answers where something came from and whether it is authentic. Carbon credits answer how much verified environmental benefit a facility created, and who owns it now."
        features={features}
      />
      <CTABanner
        title="Start with one batch"
        description="Register an organization, add a facility, and log your first checkpoint. Buying and retiring credits is free."
        primaryAction={{ label: "Create an organization", href: "/signup" }}
        secondaryAction={{ label: "See pricing", href: "/pricing" }}
      />
    </>
  );
}
