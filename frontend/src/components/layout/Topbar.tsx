import { Bell } from "lucide-react";
import Link from "next/link";
import { WalletConnectButton } from "@/components/shared/WalletConnectButton";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";

const initialsOf = (name: string) =>
  name
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part.charAt(0).toUpperCase())
    .join("");

type TopbarProps = {
  userName: string;
  unreadNotificationCount: number;
};

export function Topbar({ userName, unreadNotificationCount }: TopbarProps) {
  return (
    <header className="flex h-14 shrink-0 items-center justify-end gap-3 border-b border-neutral-200 bg-white px-6">
      <WalletConnectButton />

      <Link
        href="/notifications"
        className="relative rounded-lg p-2 text-muted-foreground hover:bg-neutral-100 hover:text-foreground"
        aria-label={
          unreadNotificationCount > 0
            ? `Notifications, ${unreadNotificationCount} unread`
            : "Notifications"
        }
      >
        <Bell className="size-4" aria-hidden />
        {unreadNotificationCount > 0 ? (
          <span className="absolute top-1 right-1 flex size-4 items-center justify-center rounded-full bg-danger-600 text-[10px] font-medium text-white tabular-nums">
            {unreadNotificationCount > 9 ? "9+" : unreadNotificationCount}
          </span>
        ) : null}
      </Link>

      <Link href="/settings/profile" aria-label="Account settings">
        <Avatar className="size-8">
          <AvatarFallback className="text-caption">
            {initialsOf(userName)}
          </AvatarFallback>
        </Avatar>
      </Link>
    </header>
  );
}
