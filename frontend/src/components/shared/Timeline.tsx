import type { ReactNode } from "react";
import type { StatusVariant } from "@/lib/status";
import { cn } from "@/lib/utils";

const markerVariants: Record<StatusVariant, string> = {
  neutral: "border-neutral-400 bg-white",
  primary: "border-primary-600 bg-primary-50",
  success: "border-success-600 bg-success-50",
  warning: "border-warning-600 bg-warning-50",
  danger: "border-danger-600 bg-danger-50",
  info: "border-info-600 bg-info-50",
};

export type TimelineEntry = {
  id: string;
  title: ReactNode;
  meta?: ReactNode | undefined;
  body?: ReactNode | undefined;
  variant?: StatusVariant | undefined;
  superseded?: boolean | undefined;
};

type TimelineProps = {
  entries: TimelineEntry[];
  className?: string | undefined;
};

export function Timeline({ entries, className }: TimelineProps) {
  return (
    <ol
      data-slot="timeline"
      className={cn("relative flex flex-col", className)}
    >
      {entries.map((entry, index) => (
        <li key={entry.id} className="relative flex gap-4 pb-6 last:pb-0">
          {index < entries.length - 1 ? (
            <span
              className="absolute top-4 left-[7px] h-full w-px bg-neutral-200"
              aria-hidden
            />
          ) : null}
          <span
            className={cn(
              "relative z-10 mt-1 size-4 shrink-0 rounded-full border-2",
              markerVariants[entry.variant ?? "neutral"],
            )}
            aria-hidden
          />
          <div
            className={cn(
              "min-w-0 flex-1 space-y-1",
              entry.superseded && "opacity-60",
            )}
          >
            <div
              className={cn(
                "font-medium",
                entry.superseded && "line-through decoration-neutral-400",
              )}
            >
              {entry.title}
            </div>
            {entry.meta ? (
              <div className="text-caption text-muted-foreground">
                {entry.meta}
              </div>
            ) : null}
            {entry.body}
          </div>
        </li>
      ))}
    </ol>
  );
}
