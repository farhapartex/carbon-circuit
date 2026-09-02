import { AppShell } from "@/components/layout/AppShell";
import { auth0 } from "@/lib/auth0";
import { listNotifications } from "@/lib/fixtures";
import { tenancyOf } from "@/lib/session";

export default async function AppLayout({ children }: LayoutProps<"/">) {
  const session = await auth0.getSession();
  const tenancy = tenancyOf(session);
  const notifications = await listNotifications();

  return (
    <AppShell
      organizationName={tenancy?.organizationName ?? "Your organization"}
      organizationType={tenancy?.organizationType ?? null}
      organizationState={tenancy?.organizationState ?? "active"}
      verificationStatus={tenancy?.verificationStatus ?? "unverified"}
      treasuryDesignated={tenancy?.isTreasuryDesignated ?? false}
      userName={session?.user.name ?? "Account"}
      userEmail={session?.user.email ?? ""}
      unreadNotificationCount={
        notifications.items.filter((notification) => !notification.read).length
      }
    >
      {children}
    </AppShell>
  );
}
