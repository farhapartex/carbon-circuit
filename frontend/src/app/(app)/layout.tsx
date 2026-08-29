import { AppShell } from "@/components/layout/AppShell";
import { getCurrentOrganization, listNotifications } from "@/lib/fixtures";
import { organizationUsers } from "@/lib/fixtures/organizations";

export default async function AppLayout({ children }: LayoutProps<"/">) {
  const organization = await getCurrentOrganization();
  const notifications = await listNotifications();
  const signedInUser = organizationUsers[0];

  return (
    <AppShell
      organization={organization}
      userName={signedInUser?.name ?? "Account"}
      unreadNotificationCount={
        notifications.items.filter((notification) => !notification.read).length
      }
    >
      {children}
    </AppShell>
  );
}
