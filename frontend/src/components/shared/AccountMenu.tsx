"use client";

import { LogOut, Settings } from "lucide-react";
import Link from "next/link";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { LOGOUT_PATH } from "@/components/shared/LogoutButton";

const initialsOf = (name: string) =>
  name
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part.charAt(0).toUpperCase())
    .join("");

type AccountMenuProps = {
  userName: string;
  userEmail: string;
};

export function AccountMenu({ userName, userEmail }: AccountMenuProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        aria-label="Account menu"
        className="rounded-full focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
      >
        <Avatar className="size-8">
          <AvatarFallback className="text-caption">
            {initialsOf(userName)}
          </AvatarFallback>
        </Avatar>
      </DropdownMenuTrigger>

      <DropdownMenuContent>
        <DropdownMenuLabel>
          <span className="font-700 block truncate text-body">{userName}</span>
          <span className="block truncate text-helper text-neutral-600">
            {userEmail}
          </span>
        </DropdownMenuLabel>

        <DropdownMenuSeparator />

        <DropdownMenuItem asChild>
          <Link href="/settings/profile">
            <Settings aria-hidden />
            Settings
          </Link>
        </DropdownMenuItem>

        <DropdownMenuItem variant="destructive" asChild>
          <a href={LOGOUT_PATH}>
            <LogOut aria-hidden />
            Sign out
          </a>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
