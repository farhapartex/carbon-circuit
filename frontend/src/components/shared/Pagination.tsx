"use client";

import { ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { CursorMeta, PageMeta } from "@/lib/types";
import { cn } from "@/lib/utils";

type OffsetPaginationProps = {
  mode: "offset";
  meta: PageMeta;
  onPageChange: (page: number) => void;
  className?: string | undefined;
};

type CursorPaginationProps = {
  mode: "cursor";
  meta: CursorMeta;
  loadedCount: number;
  onLoadMore: (cursor: string) => void;
  className?: string | undefined;
};

type PaginationProps = OffsetPaginationProps | CursorPaginationProps;

export function Pagination(props: PaginationProps) {
  if (props.mode === "cursor") {
    const { meta, loadedCount, onLoadMore, className } = props;
    return (
      <nav
        aria-label="Pagination"
        className={cn("flex items-center justify-between gap-4", className)}
      >
        <p className="text-caption text-muted-foreground">
          Showing {loadedCount} {loadedCount === 1 ? "record" : "records"}
        </p>
        {meta.hasMore && meta.nextCursor ? (
          <Button
            variant="outline"
            size="sm"
            onClick={() => onLoadMore(meta.nextCursor as string)}
          >
            Load more
          </Button>
        ) : null}
      </nav>
    );
  }

  const { meta, onPageChange, className } = props;
  const firstItem = (meta.page - 1) * meta.perPage + 1;
  const lastItem = Math.min(meta.page * meta.perPage, meta.totalItems);

  return (
    <nav
      aria-label="Pagination"
      className={cn("flex items-center justify-between gap-4", className)}
    >
      <p className="text-caption text-muted-foreground tabular-nums">
        {meta.totalItems === 0
          ? "No records"
          : `${firstItem}–${lastItem} of ${meta.totalItems}`}
      </p>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          disabled={meta.page <= 1}
          onClick={() => onPageChange(meta.page - 1)}
        >
          <ChevronLeft className="size-3.5" aria-hidden />
          Previous
        </Button>
        <span className="text-caption text-muted-foreground tabular-nums">
          Page {meta.page} of {meta.totalPages}
        </span>
        <Button
          variant="outline"
          size="sm"
          disabled={meta.page >= meta.totalPages}
          onClick={() => onPageChange(meta.page + 1)}
        >
          Next
          <ChevronRight className="size-3.5" aria-hidden />
        </Button>
      </div>
    </nav>
  );
}
