import { Check, RotateCcw } from "lucide-react";
import type { ClaimStatus } from "@/lib/status";
import { cn } from "@/lib/utils";

const STEPS = ["Submitted", "AI review", "Human review", "Decided"] as const;

const reachedStep: Record<ClaimStatus, number> = {
  draft: -1,
  submitted: 0,
  under_ai_review: 1,
  under_human_review: 2,
  more_information_requested: 2,
  approved: 3,
  rejected: 3,
};

export function ClaimStatusStepper({ status }: { status: ClaimStatus }) {
  const reached = reachedStep[status];
  const returned = status === "more_information_requested";

  if (reached < 0) {
    return (
      <p className="rounded-lg border border-dashed border-neutral-200 bg-white px-4 py-3 text-caption text-neutral-600">
        This claim is still a draft and has not entered review.
      </p>
    );
  }

  return (
    <ol className="flex flex-wrap items-center gap-x-2 gap-y-3">
      {STEPS.map((label, index) => {
        const done = index < reached;
        const active = index === reached;

        return (
          <li key={label} className="flex items-center gap-2">
            <span
              aria-current={active ? "step" : undefined}
              className={cn(
                "flex size-6 shrink-0 items-center justify-center rounded-full text-caption font-medium tabular-nums",
                done && "bg-primary-700 text-white",
                active &&
                  !returned &&
                  "bg-primary-50 text-primary-800 ring-2 ring-primary-600",
                active &&
                  returned &&
                  "bg-warning-50 text-warning-700 ring-2 ring-warning-600",
                !done && !active && "bg-neutral-100 text-neutral-600",
              )}
            >
              {done ? (
                <Check className="size-3.5" aria-hidden />
              ) : active && returned ? (
                <RotateCcw className="size-3.5" aria-hidden />
              ) : (
                index + 1
              )}
            </span>
            <span
              className={cn(
                "text-caption",
                active ? "font-medium text-neutral-900" : "text-neutral-600",
              )}
            >
              {label}
            </span>
            {index < STEPS.length - 1 ? (
              <span
                className="ml-2 hidden h-px w-8 bg-neutral-200 sm:block"
                aria-hidden
              />
            ) : null}
          </li>
        );
      })}
    </ol>
  );
}
