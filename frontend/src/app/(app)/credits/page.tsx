import type { Metadata } from "next";
import { CreditClassBreakdownChart } from "@/components/features/credits/CreditClassBreakdownChart";
import { CreditClassTable } from "@/components/features/credits/CreditClassTable";
import { CreditAmountDisplay } from "@/components/shared/CreditAmountDisplay";
import { MetricCard } from "@/components/shared/MetricCard";
import { PageHeader } from "@/components/shared/PageHeader";
import { getCreditPortfolio } from "@/lib/fixtures";

export const metadata: Metadata = { title: "Credits" };

export default async function CreditsPage() {
  const portfolio = await getCreditPortfolio();

  return (
    <>
      <PageHeader
        title="Carbon credits"
        description="Every credit permanently carries its originating facility, vintage, and activity type. Sales and retirements always operate on one specific class."
      />

      <div className="grid gap-4 sm:grid-cols-3">
        <MetricCard
          label="Available across all classes"
          value={<CreditAmountDisplay amount={portfolio.totalAvailable} />}
          hint="An aggregate for reference only — not a spendable balance."
        />
        <MetricCard
          label="Escrowed on active listings"
          value={<CreditAmountDisplay amount={portfolio.totalEscrowed} />}
          hint="Held against listings until they fill, expire, or are cancelled."
        />
        <MetricCard
          label="Retired"
          value={<CreditAmountDisplay amount={portfolio.totalRetired} />}
          hint="Permanently withdrawn and publicly recorded."
        />
      </div>

      {portfolio.balances.length > 0 ? (
        <CreditClassBreakdownChart balances={portfolio.balances} />
      ) : null}

      <CreditClassTable balances={portfolio.balances} />
    </>
  );
}
