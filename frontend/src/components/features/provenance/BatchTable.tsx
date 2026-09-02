"use client";

import { Boxes } from "lucide-react";
import Link from "next/link";
import { useState, useTransition } from "react";
import { DataTable, type DataTableColumn } from "@/components/shared/DataTable";
import { EmptyState } from "@/components/shared/EmptyState";
import { Pagination } from "@/components/shared/Pagination";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { loadMoreBatches } from "@/lib/actions/batches";
import type { Batch, CursorMeta } from "@/lib/types";

const numberFormat = new Intl.NumberFormat("en-US");

const columns: DataTableColumn<Batch>[] = [
  {
    key: "reference",
    header: "Batch",
    render: (batch) => (
      <span className="block">
        <Link
          href={`/batches/${batch.id}`}
          className="font-medium underline-offset-4 hover:underline"
        >
          {batch.componentType}
        </Link>
        <span className="block text-caption text-neutral-600">
          {batch.lotNumber ?? batch.publicReference}
        </span>
      </span>
    ),
  },
  {
    key: "facility",
    header: "Originating facility",
    render: (batch) => batch.originatingFacilityName,
  },
  {
    key: "quantity",
    header: "Quantity",
    alignEnd: true,
    render: (batch) => (
      <span className="tabular-nums">
        {numberFormat.format(batch.quantity)} {batch.unit}
      </span>
    ),
  },
  {
    key: "checkpoints",
    header: "Checkpoints",
    alignEnd: true,
    render: (batch) => (
      <span className="tabular-nums">{batch.checkpointCount}</span>
    ),
  },
  {
    key: "score",
    header: "Provenance",
    alignEnd: true,
    render: (batch) => (
      <Badge variant="outline" className="tabular-nums">
        {batch.provenanceScore.total}
      </Badge>
    ),
  },
  {
    key: "produced",
    header: "Produced",
    hideOnCard: true,
    render: (batch) => <TimestampDisplay value={batch.producedAt} dateOnly />,
  },
];

type BatchTableProps = {
  batches: Batch[];
  meta: CursorMeta;
};

export function BatchTable({ batches, meta }: BatchTableProps) {
  const [loaded, setLoaded] = useState(batches);
  const [cursor, setCursor] = useState(meta);
  const [pending, startTransition] = useTransition();

  const loadMore = (after: string) => {
    startTransition(async () => {
      const page = await loadMoreBatches(after);
      setLoaded((current) => [...current, ...page.items]);
      setCursor(page.meta);
    });
  };

  return (
    <div className="space-y-4">
      <DataTable
        columns={columns}
        rows={loaded}
        rowKey={(batch) => batch.id}
        caption="Batches registered by your organization"
        emptyState={
          <EmptyState
            icon={Boxes}
            title="No batches yet"
            description="A batch records a produced quantity of a component. Register one to start tracking where it goes."
            action={
              <Button asChild>
                <Link href="/batches/new">Register a batch</Link>
              </Button>
            }
          />
        }
      />

      {cursor.hasMore && !pending ? (
        <Pagination
          mode="cursor"
          meta={cursor}
          loadedCount={loaded.length}
          onLoadMore={loadMore}
        />
      ) : null}
    </div>
  );
}
