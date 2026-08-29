import type { LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils";

export type Feature = {
  icon: LucideIcon;
  title: string;
  description: string;
};

function FeatureCard({ icon: Icon, title, description }: Feature) {
  return (
    <article className="space-y-3 rounded-lg border border-neutral-200 bg-white p-6 shadow-sm">
      <span className="flex size-9 items-center justify-center rounded-lg bg-primary-50">
        <Icon className="size-4.5 text-primary-600" aria-hidden />
      </span>
      <h3 className="font-medium">{title}</h3>
      <p className="text-caption text-pretty text-neutral-600">{description}</p>
    </article>
  );
}

type FeatureGridProps = {
  heading: string;
  description?: string | undefined;
  features: Feature[];
  className?: string | undefined;
};

export function FeatureGrid({
  heading,
  description,
  features,
  className,
}: FeatureGridProps) {
  return (
    <section className={cn("mx-auto max-w-6xl px-6 py-20", className)}>
      <div className="max-w-2xl space-y-3">
        <h2 className="text-section-heading">{heading}</h2>
        {description ? (
          <p className="text-body text-pretty text-neutral-600">
            {description}
          </p>
        ) : null}
      </div>
      <div className="mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {features.map((feature) => (
          <FeatureCard key={feature.title} {...feature} />
        ))}
      </div>
    </section>
  );
}
