"use client";

import {
  Building2,
  CreditCard,
  KeyRound,
  UserRound,
  Users,
  Wallet,
  type LucideIcon,
} from "lucide-react";
import Link from "next/link";
import type { Route } from "next";
import { usePathname } from "next/navigation";
import { isActiveRoute } from "@/lib/navigation";
import { cn } from "@/lib/utils";

type SettingsNavItem = {
  label: string;
  description: string;
  href: Route;
  icon: LucideIcon;
};

const personalItems: SettingsNavItem[] = [
  {
    label: "Profile",
    description: "Your name, email, security, and personal wallet",
    href: "/settings/profile",
    icon: UserRound,
  },
];

const organizationItems: SettingsNavItem[] = [
  {
    label: "Organization",
    description: "Company details, verification, and product categories",
    href: "/settings/organization",
    icon: Building2,
  },
  {
    label: "Members",
    description: "Who can act for this organization",
    href: "/settings/members",
    icon: Users,
  },
  {
    label: "Treasury Address",
    description: "The address that holds your credits",
    href: "/settings/wallet",
    icon: Wallet,
  },
  {
    label: "API keys",
    description: "Keys your systems use to submit data",
    href: "/settings/api-keys",
    icon: KeyRound,
  },
  {
    label: "Billing",
    description: "Plan, usage, invoices, and payment method",
    href: "/settings/billing",
    icon: CreditCard,
  },
];

const groups = [
  { label: "Personal", items: personalItems },
  { label: "Organization", items: organizationItems },
];

export function SettingsNav() {
  const pathname = usePathname();

  return (
    <nav aria-label="Settings" className="flex shrink-0 flex-col gap-6 lg:w-56">
      {groups.map((group) => (
        <div key={group.label} className="space-y-1">
          <p className="px-3 text-caption font-medium tracking-wide text-neutral-400 uppercase">
            {group.label}
          </p>
          <ul className="flex gap-1 overflow-x-auto lg:flex-col lg:overflow-visible">
            {group.items.map((item) => {
              const active = isActiveRoute(pathname, item.href);
              return (
                <li key={item.href}>
                  <Link
                    href={item.href}
                    aria-current={active ? "page" : undefined}
                    title={item.description}
                    className={cn(
                      "flex items-center gap-2 rounded-lg px-3 py-2 text-caption font-medium whitespace-nowrap transition-colors",
                      active
                        ? "bg-primary-50 text-primary-800"
                        : "text-neutral-600 hover:bg-neutral-100 hover:text-neutral-900",
                    )}
                  >
                    <item.icon className="size-4 shrink-0" aria-hidden />
                    {item.label}
                  </Link>
                </li>
              );
            })}
          </ul>
        </div>
      ))}
    </nav>
  );
}
