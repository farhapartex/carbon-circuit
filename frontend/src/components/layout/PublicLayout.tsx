import { Leaf } from "lucide-react";
import Link from "next/link";
import type { ReactNode } from "react";

export function PublicLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col">
      <header className="border-b border-neutral-200 bg-white">
        <div className="mx-auto flex h-14 max-w-5xl items-center gap-2 px-6">
          <Leaf className="size-4 text-primary-700" aria-hidden />
          <Link href="/" className="rounded-sm font-semibold">
            CarbonCircuit
          </Link>
        </div>
      </header>
      <main className="flex-1">
        <div className="mx-auto max-w-5xl px-6 py-10">{children}</div>
      </main>
      <footer className="border-t border-neutral-200 bg-white">
        <div className="mx-auto flex max-w-5xl flex-wrap items-center gap-x-6 gap-y-2 px-6 py-6 text-caption text-muted-foreground">
          <span>Verified provenance and carbon credits.</span>
          <Link
            href="/marketplace/retirements"
            className="rounded-sm underline underline-offset-4"
          >
            Public retirement log
          </Link>
          <Link
            href="/pricing"
            className="rounded-sm underline underline-offset-4"
          >
            Pricing
          </Link>
        </div>
      </footer>
    </div>
  );
}
