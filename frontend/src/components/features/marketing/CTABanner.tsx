import Link from "next/link";
import type { MarketingAction } from "@/components/features/marketing/HeroSection";
import { Button } from "@/components/ui/button";

type CTABannerProps = {
  title: string;
  description: string;
  primaryAction: MarketingAction;
  secondaryAction?: MarketingAction | undefined;
};

export function CTABanner({
  title,
  description,
  primaryAction,
  secondaryAction,
}: CTABannerProps) {
  return (
    <section className="mx-auto max-w-6xl px-6 py-16">
      <div className="flex flex-col gap-6 rounded-lg border border-neutral-200 bg-primary-50 px-8 py-10 md:flex-row md:items-center md:justify-between">
        <div className="max-w-xl space-y-2">
          <h2 className="text-section-heading text-primary-800">{title}</h2>
          <p className="text-body text-pretty text-primary-800/80">
            {description}
          </p>
        </div>
        <div className="flex shrink-0 flex-wrap gap-3">
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
    </section>
  );
}
