"use client";

import { X } from "lucide-react";
import type { ReactNode } from "react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type FilterBarProps = {
  children: ReactNode;
  activeCount: number;
  onClear: () => void;
  className?: string | undefined;
};

export function FilterBar({
  children,
  activeCount,
  onClear,
  className,
}: FilterBarProps) {
  return (
    <div
      data-slot="filter-bar"
      role="search"
      className={cn(
        "flex flex-wrap items-center gap-2 rounded-lg border border-neutral-200 bg-white p-3",
        className,
      )}
    >
      {children}
      {activeCount > 0 ? (
        <Button variant="ghost" size="sm" onClick={onClear} className="ml-auto">
          <X className="size-3.5" aria-hidden />
          Clear {activeCount} filter{activeCount === 1 ? "" : "s"}
        </Button>
      ) : null}
    </div>
  );
}
