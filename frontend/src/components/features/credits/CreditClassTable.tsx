"use client";

import { Coins } from "lucide-react";
import Link from "next/link";
import { CreditAmountDisplay } from "@/components/shared/CreditAmountDisplay";
import { DataTable, type DataTableColumn } from "@/components/shared/DataTable";
import { EmptyState } from "@/components/shared/EmptyState";
import { Button } from "@/components/ui/button";
import { countryName } from "@/lib/countries";
import { activityTypeLabels } from "@/lib/labels";
import { isZeroCreditAmount } from "@/lib/decimal";
import type { CreditClassBalance } from "@/lib/types";

const columns: DataTableColumn<CreditClassBalance>[] = [
  {
    key: "class",
    header: "Credit class",
    render: (balance) => (
      <span className="block">
        <span className="block font-medium">
          {activityTypeLabels[balance.creditClass.activityType]}
        </span>
        <span className="block text-caption text-neutral-600">
          {balance.creditClass.facilityName} ·{" "}
          {countryName(balance.creditClass.facilityCountry)} · vintage{" "}
          {balance.creditClass.vintageYear}
        </span>
      </span>
    ),
  },
  {
    key: "available",
    header: "Available",
    alignEnd: true,
    render: (balance) => <CreditAmountDisplay amount={balance.available} />,
  },
  {
    key: "escrowed",
    header: "Escrowed",
    alignEnd: true,
    render: (balance) => <CreditAmountDisplay amount={balance.escrowed} />,
  },
  {
    key: "retired",
    header: "Retired",
    alignEnd: true,
    render: (balance) => <CreditAmountDisplay amount={balance.retired} />,
  },
  {
    key: "actions",
    header: "Actions",
    alignEnd: true,
    render: (balance) => {
      const spendable = !isZeroCreditAmount(balance.available);

      if (!spendable) {
        return (
          <span className="text-caption text-neutral-600">
            Nothing available
          </span>
        );
      }

      return (
        <span className="flex flex-wrap justify-end gap-2">
          <Button asChild size="sm" variant="outline">
            <Link href="/marketplace/my-listings/new">Sell</Link>
          </Button>
          <Button asChild size="sm" variant="outline">
            <Link href="/marketplace/my-retirements">Retire</Link>
          </Button>
        </span>
      );
    },
  },
];

export function CreditClassTable({
  balances,
}: {
  balances: CreditClassBalance[];
}) {
  return (
    <DataTable
      columns={columns}
      rows={balances}
      rowKey={(balance) => balance.creditClass.tokenId}
      caption="Carbon credits held by your organization, by credit class"
      emptyState={
        <EmptyState
          icon={Coins}
          title="No credits yet"
          description="Credits arrive when a verifier approves a sustainability claim for one of your facilities. Each one permanently carries its originating facility, vintage, and activity type."
          action={
            <Button asChild>
              <Link href="/claims/new">Submit a claim</Link>
            </Button>
          }
        />
      }
    />
  );
}
