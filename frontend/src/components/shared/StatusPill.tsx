import { cva } from "class-variance-authority";
import type { StatusPresentation, StatusVariant } from "@/lib/status";
import { cn } from "@/lib/utils";

const pillVariants = cva(
  "inline-flex w-fit shrink-0 items-center gap-1.5 rounded-full px-2.5 py-0.5 text-caption font-medium whitespace-nowrap",
  {
    variants: {
      variant: {
        neutral: "bg-neutral-100 text-neutral-600",
        primary: "bg-primary-50 text-primary-800",
        success: "bg-success-50 text-success-700",
        warning: "bg-warning-50 text-warning-700",
        danger: "bg-danger-50 text-danger-700",
        info: "bg-info-50 text-info-700",
      },
    },
    defaultVariants: { variant: "neutral" },
  },
);

const dotVariants = cva("size-1.5 shrink-0 rounded-full", {
  variants: {
    variant: {
      neutral: "bg-neutral-400",
      primary: "bg-primary-600",
      success: "bg-success-600",
      warning: "bg-warning-600",
      danger: "bg-danger-600",
      info: "bg-info-600",
    },
  },
  defaultVariants: { variant: "neutral" },
});

type StatusPillProps = {
  presentation: StatusPresentation;
  showDot?: boolean | undefined;
  className?: string | undefined;
};

export function StatusPill({
  presentation,
  showDot = true,
  className,
}: StatusPillProps) {
  const variant: StatusVariant = presentation.variant;

  return (
    <span
      data-slot="status-pill"
      data-variant={variant}
      className={cn(pillVariants({ variant }), className)}
    >
      {showDot ? (
        <span className={dotVariants({ variant })} aria-hidden />
      ) : null}
      {presentation.label}
    </span>
  );
}
