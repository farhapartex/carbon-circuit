"use client";

import { Leaf } from "lucide-react";
import { CreditAmountDisplay } from "@/components/shared/CreditAmountDisplay";
import { DataTable, type DataTableColumn } from "@/components/shared/DataTable";
import { EmptyState } from "@/components/shared/EmptyState";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { activityTypeLabels } from "@/lib/labels";
import { explorerTransactionUrl } from "@/lib/wallet/chain";
import type { Retirement } from "@/lib/types";

const columns: DataTableColumn<Retirement>[] = [
  {
    key: "class",
    header: "Credit class",
    render: (retirement) => (
      <span className="block">
        <span className="block font-medium">
          {activityTypeLabels[retirement.creditClass.activityType]}
        </span>
        <span className="block text-caption text-neutral-600">
          {retirement.creditClass.facilityName} · vintage{" "}
          {retirement.creditClass.vintageYear}
        </span>
      </span>
    ),
  },
  {
    key: "quantity",
    header: "Quantity",
    alignEnd: true,
    render: (retirement) => (
      <CreditAmountDisplay amount={retirement.quantity} />
    ),
  },
  {
    key: "purpose",
    header: "Purpose",
    render: (retirement) => (
      <span className="text-pretty">{retirement.purpose}</span>
    ),
  },
  {
    key: "retiredAt",
    header: "Retired",
    render: (retirement) => (
      <TimestampDisplay value={retirement.retiredAt} dateOnly />
    ),
  },
  {
    key: "transaction",
    header: "On chain",
    hideOnCard: true,
    render: (retirement) =>
      retirement.transactionHash ? (
        <a
          href={explorerTransactionUrl(retirement.transactionHash)}
          target="_blank"
          rel="noreferrer noopener"
          className="text-helper underline-offset-4 hover:underline"
        >
          View
        </a>
      ) : (
        <span className="text-caption text-neutral-600">Pending</span>
      ),
  },
];

export function MyRetirementsTable({
  retirements,
}: {
  retirements: Retirement[];
}) {
  return (
    <DataTable
      columns={columns}
      rows={retirements}
      rowKey={(retirement) => retirement.id}
      caption="Credits your organization has retired"
      emptyState={
        <EmptyState
          icon={Leaf}
          title="No retirements yet"
          description="Retiring a credit permanently withdraws it from circulation to claim an offset. Every retirement is published on the public log."
          action={null}
        />
      }
    />
  );
}
