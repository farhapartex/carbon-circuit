import { Leaf } from "lucide-react";
import Link from "next/link";
import { LogoutButton } from "@/components/shared/LogoutButton";
import { getSignedInUser } from "@/lib/fixtures";

export default async function OnboardingLayout({ children }: LayoutProps<"/">) {
  const user = await getSignedInUser();

  return (
    <div className="flex min-h-screen flex-col">
      <header className="border-b border-neutral-200 bg-white">
        <div className="mx-auto flex h-14 max-w-3xl items-center gap-3 px-6">
          <Link href="/" className="flex items-center gap-2 rounded-sm">
            <Leaf className="size-4 text-primary-700" aria-hidden />
            <span className="font-semibold">CarbonCircuit</span>
          </Link>
          <span className="ml-auto hidden text-helper text-neutral-600 sm:block">
            Signed in as {user.email}
          </span>
          <LogoutButton />
        </div>
      </header>
      <main className="flex-1">
        <div className="mx-auto max-w-3xl space-y-8 px-6 py-10">{children}</div>
      </main>
    </div>
  );
}
