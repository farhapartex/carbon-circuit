import type { Metadata } from "next";
import { ExternalLink } from "lucide-react";
import { RetirementLogTable } from "@/components/features/marketplace/RetirementLogTable";
import { listRetirements } from "@/lib/fixtures";

export const metadata: Metadata = {
  title: "Retirement log",
  description:
    "Every carbon credit permanently retired on CarbonCircuit, open to anyone.",
};

export default async function RetirementLogPage() {
  const retirements = await listRetirements();

  return (
    <div className="space-y-6">
      <header className="max-w-3xl space-y-3">
        <h1 className="text-page-title">Public retirement log</h1>
        <p className="text-body text-pretty text-neutral-600">
          A retired credit is permanently removed from circulation and can never
          be resold, retraded, or retired again by anyone. This log exists so
          that claim is checkable by a third party rather than taken on trust.
          Every entry names the facility that earned the credit, the year the
          reduction happened, and the practice behind it.
        </p>
        <p className="inline-flex items-center gap-1.5 text-caption text-neutral-600">
          <ExternalLink className="size-3" aria-hidden />
          Each row links to the on-chain transaction that burned the credit.
        </p>
      </header>

      <RetirementLogTable retirements={retirements.items} />
    </div>
  );
}
