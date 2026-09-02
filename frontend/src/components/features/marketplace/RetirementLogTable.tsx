"use client";

import { ExternalLink } from "lucide-react";
import { CreditAmountDisplay } from "@/components/shared/CreditAmountDisplay";
import { DataTable, type DataTableColumn } from "@/components/shared/DataTable";
import { EmptyState } from "@/components/shared/EmptyState";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import type { Retirement } from "@/lib/types";
import { activityTypeLabels } from "@/lib/labels";
import { explorerTransactionUrl } from "@/lib/wallet/chain";

const columns: DataTableColumn<Retirement>[] = [
  {
    key: "retiringOrganization",
    header: "Retired by",
    render: (retirement) => retirement.retiringOrganizationName,
  },
  {
    key: "facility",
    header: "Originating facility",
    render: (retirement) => (
      <span>
        {retirement.creditClass.facilityName}
        <span className="block text-caption text-neutral-600">
          {retirement.creditClass.facilityCountry}
        </span>
      </span>
    ),
  },
  {
    key: "vintage",
    header: "Vintage",
    render: (retirement) => retirement.creditClass.vintageYear,
  },
  {
    key: "activityType",
    header: "Activity",
    render: (retirement) =>
      activityTypeLabels[retirement.creditClass.activityType],
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
    header: "Stated purpose",
    render: (retirement) => (
      <span className="text-neutral-600">{retirement.purpose}</span>
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
    header: "Proof",
    render: (retirement) =>
      retirement.transactionHash ? (
        <a
          href={explorerTransactionUrl(retirement.transactionHash)}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-1 rounded-sm text-primary-700 underline underline-offset-4"
        >
          On-chain
          <ExternalLink className="size-3" aria-hidden />
        </a>
      ) : (
        <span className="text-neutral-600">Pending</span>
      ),
  },
];

export function RetirementLogTable({
  retirements,
}: {
  retirements: Retirement[];
}) {
  return (
    <DataTable
      caption="Public retirement log"
      columns={columns}
      rows={retirements}
      rowKey={(retirement) => retirement.id}
      emptyState={
        <EmptyState
          title="No credits retired yet"
          description="When an organization retires a credit, it appears here permanently, naming the facility that earned it."
          action={<span />}
        />
      }
    />
  );
}
