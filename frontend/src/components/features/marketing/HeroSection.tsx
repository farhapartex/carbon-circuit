import Link from "next/link";
import type { Route } from "next";
import type { ReactNode } from "react";
import { Button } from "@/components/ui/button";

export type MarketingAction = {
  label: string;
  href: Route;
};

type HeroSectionProps = {
  eyebrow?: string | undefined;
  title: string;
  description: string;
  primaryAction: MarketingAction;
  secondaryAction?: MarketingAction | undefined;
  aside?: ReactNode | undefined;
};

export function HeroSection({
  eyebrow,
  title,
  description,
  primaryAction,
  secondaryAction,
  aside,
}: HeroSectionProps) {
  return (
    <section className="border-b border-neutral-200 bg-white">
      <div className="mx-auto grid max-w-6xl gap-12 px-6 py-20 lg:grid-cols-[1.1fr_1fr] lg:items-center">
        <div className="space-y-6">
          {eyebrow ? (
            <p className="text-caption font-medium tracking-wide text-primary-700 uppercase">
              {eyebrow}
            </p>
          ) : null}
          <h1 className="text-page-title text-balance lg:text-5xl lg:leading-tight">
            {title}
          </h1>
          <p className="max-w-xl text-lg text-pretty text-neutral-600">
            {description}
          </p>
          <div className="flex flex-wrap gap-3">
            <Button asChild size="lg">
              <Link href={primaryAction.href}>{primaryAction.label}</Link>
            </Button>
            {secondaryAction ? (
              <Button asChild size="lg" variant="outline">
                <Link href={secondaryAction.href}>{secondaryAction.label}</Link>
              </Button>
            ) : null}
          </div>
        </div>
        {aside}
      </div>
    </section>
  );
}
