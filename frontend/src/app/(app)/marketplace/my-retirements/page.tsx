import type { Metadata } from "next";
import { MyRetirementsTable } from "@/components/features/marketplace/MyRetirementsTable";
import { RetirementForm } from "@/components/features/marketplace/RetirementForm";
import { PageHeader } from "@/components/shared/PageHeader";
import { getCreditPortfolio, listRetirements } from "@/lib/fixtures";

export const metadata: Metadata = { title: "My retirements" };

export default async function MyRetirementsPage() {
  const [retirements, portfolio] = await Promise.all([
    listRetirements(),
    getCreditPortfolio(),
  ]);

  return (
    <>
      <PageHeader
        backTo={{ href: "/marketplace", label: "Marketplace" }}
        title="My retirements"
        description="Retirement is permanent and public. It is the mechanism that stops the same environmental benefit being claimed twice."
        actions={<RetirementForm balances={portfolio.balances} />}
      />

      <MyRetirementsTable retirements={retirements.items} />
    </>
  );
}
