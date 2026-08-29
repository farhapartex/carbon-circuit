import type { LucideIcon } from "lucide-react";
import { Inbox } from "lucide-react";
import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

type EmptyStateProps = {
  title: string;
  description: string;
  action: ReactNode;
  icon?: LucideIcon | undefined;
  className?: string | undefined;
};

export function EmptyState({
  title,
  description,
  action,
  icon: Icon = Inbox,
  className,
}: EmptyStateProps) {
  return (
    <div
      data-slot="empty-state"
      className={cn(
        "flex flex-col items-center gap-4 rounded-lg border border-dashed border-neutral-200 bg-white px-6 py-14 text-center",
        className,
      )}
    >
      <span className="flex size-10 items-center justify-center rounded-full bg-primary-50">
        <Icon className="size-5 text-primary-600" aria-hidden />
      </span>
      <div className="space-y-1">
        <p className="font-medium">{title}</p>
        <p className="mx-auto max-w-sm text-caption text-muted-foreground">
          {description}
        </p>
      </div>
      {action}
    </div>
  );
}
