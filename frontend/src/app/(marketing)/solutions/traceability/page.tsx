import type { Metadata } from "next";
import {
  Boxes,
  GitBranch,
  Gauge,
  PlugZap,
  ScanLine,
  ShieldCheck,
} from "lucide-react";
import { CTABanner } from "@/components/features/marketing/CTABanner";
import { FeatureGrid } from "@/components/features/marketing/FeatureGrid";
import { HeroSection } from "@/components/features/marketing/HeroSection";
import { HowItWorksSteps } from "@/components/features/marketing/HowItWorksSteps";
import { IndustryBand } from "@/components/features/marketing/IndustryBand";

export const metadata: Metadata = {
  title: "Traceability",
  description:
    "Record the verifiable journey of a batch through every facility that handles it.",
};

const features = [
  {
    icon: Boxes,
    title: "Batch tracking",
    description:
      "A batch is a defined quantity produced together. It carries its Product Category for life, so claim types, checkpoint expectations, and presentation stay unambiguous.",
  },
  {
    icon: ScanLine,
    title: "Checkpoint history",
    description:
      "Any party in the chain of custody sees the full history, not only the legs they logged. Corrections supersede rather than overwrite, and both records stay visible.",
  },
  {
    icon: Gauge,
    title: "A Provenance Score that explains itself",
    description:
      "Completeness, anchoring, chain depth, reporting timeliness, and the facility's sustainability record. Each component shows what it contributed, so a low score tells you what to fix.",
  },
  {
    icon: GitBranch,
    title: "Multi-level component chains",
    description:
      "A finished device references the PCB batch, which references the wafer batch, up to ten levels deep. Downstream parties see one level up, never a competitor's whole sourcing network.",
  },
  {
    icon: ShieldCheck,
    title: "Anomaly detection built in",
    description:
      "Impossible travel, backdated and future-dated events, and velocity spikes are flagged with the computed evidence attached, not buried in a log nobody reads.",
  },
  {
    icon: PlugZap,
    title: "Portal or API, same records",
    description:
      "Submit through the dashboard or straight from your ERP. Every API record carries your own external identifier, so replaying a day after an outage creates no duplicates.",
  },
];

const steps = [
  {
    title: "Register a batch",
    description:
      "Choose the Product Category, enter the component and quantity, and link any parent batches it was built from.",
  },
  {
    title: "Log checkpoints as it moves",
    description:
      "Whoever handles that leg records the event type, location, and time. Late or impossible entries are flagged rather than silently accepted.",
  },
  {
    title: "Anchor to the ledger",
    description:
      "Every ten minutes, one Merkle root covering the whole platform is written on-chain. Cost does not grow with the number of customers.",
  },
  {
    title: "Hand someone the proof",
    description:
      "A QR code resolves to a public page with the timeline, the score, and an inclusion proof anyone can verify themselves.",
  },
];

export default function TraceabilityPage() {
  return (
    <>
      <HeroSection
        eyebrow="For manufacturers, assemblers, and logistics partners"
        title="A supply chain record nobody can quietly rewrite"
        description="Electronics passes through five to eight facilities across several countries, and most origin claims cannot be checked by anyone downstream. CarbonCircuit records each step as it happens and anchors it where it cannot be edited after the fact."
        primaryAction={{ label: "Create an organization", href: "/signup" }}
        secondaryAction={{ label: "Compare plans", href: "/pricing" }}
      />
      <IndustryBand />
      <FeatureGrid
        heading="What you get"
        description="Electronics is the launch vertical, chosen because conflict minerals, e-waste, and opaque multi-tier sourcing are real and well documented. The model is not specific to it."
        features={features}
      />
      <HowItWorksSteps heading="How it works" steps={steps} />
      <CTABanner
        title="Track your first batch today"
        description="Starter includes 50 batches and 500 checkpoints a month, with manual entry through the portal."
        primaryAction={{ label: "Get started", href: "/signup" }}
        secondaryAction={{
          label: "See carbon credits",
          href: "/solutions/carbon-credits",
        }}
      />
    </>
  );
}
