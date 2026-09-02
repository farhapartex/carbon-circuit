import { ArrowLeft } from "lucide-react";
import type { Route } from "next";
import Link from "next/link";
import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

type BackDestination = {
  href: Route;
  label: string;
};

type PageHeaderProps = {
  title: string;
  backTo?: BackDestination | undefined;
  description?: string | undefined;
  actions?: ReactNode | undefined;
  meta?: ReactNode | undefined;
  className?: string | undefined;
};

export function PageHeader({
  title,
  backTo,
  description,
  actions,
  meta,
  className,
}: PageHeaderProps) {
  return (
    <header
      data-slot="page-header"
      className={cn(
        "flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between",
        className,
      )}
    >
      <div className="space-y-2">
        {backTo ? (
          <Link
            href={backTo.href}
            className="inline-flex items-center gap-1.5 text-caption text-muted-foreground underline-offset-4 hover:text-foreground hover:underline"
          >
            <ArrowLeft className="size-3.5" aria-hidden />
            {backTo.label}
          </Link>
        ) : null}
        <div className="flex flex-wrap items-center gap-3">
          <h1 className="text-page-title">{title}</h1>
          {meta}
        </div>
        {description ? (
          <p className="max-w-2xl text-body text-muted-foreground">
            {description}
          </p>
        ) : null}
      </div>
      {actions ? (
        <div className="flex shrink-0 flex-wrap items-center gap-2">
          {actions}
        </div>
      ) : null}
    </header>
  );
}
