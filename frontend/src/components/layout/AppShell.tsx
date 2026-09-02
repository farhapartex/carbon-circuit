import type { ReactNode } from "react";
import { Sidebar } from "@/components/layout/Sidebar";
import { Topbar } from "@/components/layout/Topbar";
import { VerificationStatusBanner } from "@/components/layout/VerificationStatusBanner";
import { WalletProvider } from "@/components/shared/WalletProvider";
import type { VerificationStatus } from "@/lib/status";
import type {
  OrganizationState,
  OrganizationType,
} from "@/lib/types/organization";

type AppShellProps = {
  organizationName: string;
  organizationType: OrganizationType | null;
  organizationState: OrganizationState;
  verificationStatus: VerificationStatus;
  userName: string;
  userEmail: string;
  unreadNotificationCount: number;
  children: ReactNode;
};

export function AppShell({
  organizationName,
  organizationType,
  organizationState,
  verificationStatus,
  userName,
  userEmail,
  unreadNotificationCount,
  children,
}: AppShellProps) {
  return (
    <WalletProvider>
      <div className="flex min-h-screen">
        <Sidebar
          organizationType={organizationType}
          organizationName={organizationName}
          verificationStatus={verificationStatus}
          organizationState={organizationState}
        />
        <div className="flex min-w-0 flex-1 flex-col">
          <Topbar
            userName={userName}
            userEmail={userEmail}
            unreadNotificationCount={unreadNotificationCount}
          />
          <VerificationStatusBanner
            organizationState={organizationState}
            verificationStatus={verificationStatus}
          />
          <main className="flex-1 px-6 py-8">
            <div className="mx-auto max-w-6xl space-y-8">{children}</div>
          </main>
        </div>
      </div>
    </WalletProvider>
  );
}
