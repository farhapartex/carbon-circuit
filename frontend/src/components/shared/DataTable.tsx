"use client";

import { ArrowDown, ArrowUp, ChevronsUpDown } from "lucide-react";
import type { ReactNode } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useUiContextStore } from "@/stores/ui-context";
import { cn } from "@/lib/utils";

export type SortDirection = "asc" | "desc";

export type SortState = {
  field: string;
  direction: SortDirection;
};

export const toSortParameter = (sort: SortState) =>
  sort.direction === "desc" ? `-${sort.field}` : sort.field;

export type DataTableColumn<TRow> = {
  key: string;
  header: string;
  render: (row: TRow) => ReactNode;
  sortable?: boolean | undefined;
  alignEnd?: boolean | undefined;
  hideOnCard?: boolean | undefined;
};

type DataTableProps<TRow> = {
  columns: DataTableColumn<TRow>[];
  rows: TRow[];
  rowKey: (row: TRow) => string;
  caption: string;
  sort?: SortState | undefined;
  onSortChange?: ((sort: SortState) => void) | undefined;
  emptyState?: ReactNode | undefined;
  className?: string | undefined;
};

const nextDirection = (
  column: string,
  sort: SortState | undefined,
): SortDirection =>
  sort?.field === column && sort.direction === "asc" ? "desc" : "asc";

function SortIndicator({
  active,
  direction,
}: {
  active: boolean;
  direction: SortDirection;
}) {
  if (!active) {
    return <ChevronsUpDown className="size-3.5 opacity-50" aria-hidden />;
  }
  return direction === "asc" ? (
    <ArrowUp className="size-3.5" aria-hidden />
  ) : (
    <ArrowDown className="size-3.5" aria-hidden />
  );
}

export function DataTable<TRow>({
  columns,
  rows,
  rowKey,
  caption,
  sort,
  onSortChange,
  emptyState,
  className,
}: DataTableProps<TRow>) {
  const density = useUiContextStore((state) => state.tableDensity);
  const cellPadding = density === "compact" ? "py-1.5" : "py-3";

  if (rows.length === 0 && emptyState) {
    return <>{emptyState}</>;
  }

  return (
    <div data-slot="data-table" className={className}>
      <div className="hidden md:block">
        <Table>
          <caption className="sr-only">{caption}</caption>
          <TableHeader>
            <TableRow>
              {columns.map((column) => {
                const active = sort?.field === column.key;
                return (
                  <TableHead
                    key={column.key}
                    className={cn(column.alignEnd && "text-right")}
                    aria-sort={
                      active
                        ? sort.direction === "asc"
                          ? "ascending"
                          : "descending"
                        : undefined
                    }
                  >
                    {column.sortable && onSortChange ? (
                      <button
                        type="button"
                        onClick={() =>
                          onSortChange({
                            field: column.key,
                            direction: nextDirection(column.key, sort),
                          })
                        }
                        className={cn(
                          "inline-flex items-center gap-1 rounded-sm font-medium hover:text-foreground",
                          column.alignEnd && "flex-row-reverse",
                        )}
                      >
                        {column.header}
                        <SortIndicator
                          active={active}
                          direction={sort?.direction ?? "asc"}
                        />
                      </button>
                    ) : (
                      column.header
                    )}
                  </TableHead>
                );
              })}
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row) => (
              <TableRow key={rowKey(row)}>
                {columns.map((column) => (
                  <TableCell
                    key={column.key}
                    className={cn(cellPadding, column.alignEnd && "text-right")}
                  >
                    {column.render(row)}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      <ul className="flex flex-col gap-3 md:hidden">
        {rows.map((row) => (
          <li
            key={rowKey(row)}
            className="rounded-lg border border-neutral-200 bg-white p-4 shadow-sm"
          >
            <dl className="grid gap-2">
              {columns
                .filter((column) => !column.hideOnCard)
                .map((column) => (
                  <div
                    key={column.key}
                    className="flex items-baseline justify-between gap-4"
                  >
                    <dt className="text-caption text-muted-foreground">
                      {column.header}
                    </dt>
                    <dd className="text-right">{column.render(row)}</dd>
                  </div>
                ))}
            </dl>
          </li>
        ))}
      </ul>
    </div>
  );
}
