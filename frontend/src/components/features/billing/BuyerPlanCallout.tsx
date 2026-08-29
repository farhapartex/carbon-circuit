import { Sparkles } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";

export function BuyerPlanCallout() {
  return (
    <section className="rounded-lg border border-primary-600/30 bg-primary-50 p-8">
      <div className="flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
        <div className="max-w-2xl space-y-2">
          <p className="inline-flex items-center gap-2 text-caption font-medium text-primary-800">
            <Sparkles className="size-4" aria-hidden />
            Buying and retiring credits is free
          </p>
          <h2 className="text-section-heading text-primary-800">
            Credit buyers pay nothing
          </h2>
          <p className="text-body text-pretty text-primary-800/80">
            An organization that only purchases and retires credits pays no
            subscription. Charging a company for the privilege of spending money
            on the marketplace is the fastest way to have no buyers, and a
            marketplace with no buyers is worthless to the sellers who are
            paying. Our revenue comes from the seller-side transaction fee on
            trades that would not have happened otherwise.
          </p>
        </div>
        <Button asChild size="lg" className="shrink-0">
          <Link href="/signup">Start buying</Link>
        </Button>
      </div>
    </section>
  );
}
