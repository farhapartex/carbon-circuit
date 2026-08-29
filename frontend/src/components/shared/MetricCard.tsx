import type { ReactNode } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";

type MetricCardProps = {
  label: string;
  value: ReactNode;
  hint?: string | undefined;
  footer?: ReactNode | undefined;
  className?: string | undefined;
};

export function MetricCard({
  label,
  value,
  hint,
  footer,
  className,
}: MetricCardProps) {
  return (
    <Card data-slot="metric-card" className={cn("gap-0", className)}>
      <CardContent className="space-y-1">
        <p className="text-caption text-muted-foreground">{label}</p>
        <p className="text-section-heading tabular-nums">{value}</p>
        {hint ? (
          <p className="text-caption text-muted-foreground">{hint}</p>
        ) : null}
        {footer}
      </CardContent>
    </Card>
  );
}
