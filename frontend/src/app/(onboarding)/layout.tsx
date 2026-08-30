import { Leaf } from "lucide-react";
import Link from "next/link";

export default function OnboardingLayout({ children }: LayoutProps<"/">) {
  return (
    <div className="flex min-h-screen flex-col">
      <header className="border-b border-neutral-200 bg-white">
        <div className="mx-auto flex h-14 max-w-3xl items-center gap-2 px-6">
          <Leaf className="size-4 text-primary-700" aria-hidden />
          <Link href="/" className="rounded-sm font-semibold">
            CarbonCircuit
          </Link>
        </div>
      </header>
      <main className="flex-1">
        <div className="mx-auto max-w-3xl space-y-8 px-6 py-10">{children}</div>
      </main>
    </div>
  );
}
