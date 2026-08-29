import type { ReactNode } from "react";
import { Sidebar } from "@/components/layout/Sidebar";
import { Topbar } from "@/components/layout/Topbar";
import { VerificationStatusBanner } from "@/components/layout/VerificationStatusBanner";
import { Toaster } from "@/components/shared/Toaster";
import { WalletProvider } from "@/components/shared/WalletProvider";
import type { Organization } from "@/lib/types";

type AppShellProps = {
  organization: Organization;
  userName: string;
  unreadNotificationCount: number;
  children: ReactNode;
};

export function AppShell({
  organization,
  userName,
  unreadNotificationCount,
  children,
}: AppShellProps) {
  return (
    <WalletProvider>
      <div className="flex min-h-screen">
        <Sidebar
          organizationType={organization.type}
          organizationName={organization.name}
          verificationStatus={organization.verificationStatus}
          organizationState={organization.state}
        />
        <div className="flex min-w-0 flex-1 flex-col">
          <Topbar
            userName={userName}
            unreadNotificationCount={unreadNotificationCount}
          />
          <VerificationStatusBanner organization={organization} />
          <main className="flex-1 px-6 py-8">
            <div className="mx-auto max-w-6xl space-y-8">{children}</div>
          </main>
        </div>
        <Toaster />
      </div>
    </WalletProvider>
  );
}
