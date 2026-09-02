import { Check } from "lucide-react";
import { cn } from "@/lib/utils";

type StepIndicatorProps = {
  steps: readonly string[];
  currentIndex: number;
  className?: string | undefined;
};

export function StepIndicator({
  steps,
  currentIndex,
  className,
}: StepIndicatorProps) {
  return (
    <ol
      className={cn("flex flex-wrap items-center gap-x-2 gap-y-3", className)}
    >
      {steps.map((label, index) => {
        const done = index < currentIndex;
        const active = index === currentIndex;

        return (
          <li key={label} className="flex items-center gap-2">
            <span
              aria-current={active ? "step" : undefined}
              className={cn(
                "flex size-6 shrink-0 items-center justify-center rounded-full text-caption font-medium tabular-nums",
                done && "bg-primary-700 text-white",
                active &&
                  "bg-primary-50 text-primary-800 ring-2 ring-primary-600",
                !done && !active && "bg-neutral-100 text-neutral-600",
              )}
            >
              {done ? <Check className="size-3.5" aria-hidden /> : index + 1}
            </span>
            <span
              className={cn(
                "text-caption",
                active ? "font-medium text-neutral-900" : "text-neutral-600",
              )}
            >
              {label}
            </span>
            {index < steps.length - 1 ? (
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
