import { Leaf } from "lucide-react";
import Link from "next/link";
import type { Route } from "next";
import type { ReactNode } from "react";
import { Button } from "@/components/ui/button";

const marketingLinks: { label: string; href: Route }[] = [
  { label: "Traceability", href: "/solutions/traceability" },
  { label: "Carbon credits", href: "/solutions/carbon-credits" },
  { label: "Pricing", href: "/pricing" },
  { label: "Retirement log", href: "/marketplace/retirements" },
];

export function MarketingShell({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col">
      <header className="border-b border-neutral-200 bg-white">
        <div className="mx-auto flex h-16 max-w-6xl items-center gap-8 px-6">
          <Link href="/" className="flex items-center gap-2 rounded-sm">
            <Leaf className="size-5 text-primary-700" aria-hidden />
            <span className="font-semibold">CarbonCircuit</span>
          </Link>
          <nav aria-label="Marketing" className="hidden gap-6 md:flex">
            {marketingLinks.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className="rounded-sm text-caption text-neutral-600 hover:text-neutral-900"
              >
                {link.label}
              </Link>
            ))}
          </nav>
          <div className="ml-auto flex items-center gap-2">
            <Button asChild variant="ghost" size="sm">
              <Link href="/login">Sign in</Link>
            </Button>
            <Button asChild size="sm">
              <Link href="/signup">Get started</Link>
            </Button>
          </div>
        </div>
      </header>
      <main className="flex-1">{children}</main>
      <footer className="border-t border-neutral-200 bg-white">
        <div className="mx-auto flex max-w-6xl flex-wrap items-center gap-x-6 gap-y-2 px-6 py-8 text-caption text-muted-foreground">
          <span>Provenance and carbon credits for physical supply chains.</span>
          {marketingLinks.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className="rounded-sm underline underline-offset-4"
            >
              {link.label}
            </Link>
          ))}
        </div>
      </footer>
    </div>
  );
}
