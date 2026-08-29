import { Leaf } from "lucide-react";
import Link from "next/link";
import type { ReactNode } from "react";

export function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center px-6 py-12">
      <Link href="/" className="mb-8 flex items-center gap-2 rounded-sm">
        <Leaf className="size-5 text-primary-700" aria-hidden />
        <span className="text-section-heading">CarbonCircuit</span>
      </Link>
      <div className="w-full max-w-md rounded-lg border border-neutral-200 bg-white p-8 shadow-sm">
        {children}
      </div>
    </div>
  );
}
