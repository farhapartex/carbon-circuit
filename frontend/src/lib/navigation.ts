import {
  Bell,
  Boxes,
  ClipboardCheck,
  CreditCard,
  Factory,
  FileCheck2,
  History,
  KeyRound,
  LayoutDashboard,
  Leaf,
  ListChecks,
  Settings,
  ShoppingCart,
  Store,
  Users,
  Wallet,
  type LucideIcon,
} from "lucide-react";
import type { Route } from "next";
import type {
  Organization,
  OrganizationState,
  OrganizationType,
} from "@/lib/types";
import type { VerificationStatus } from "@/lib/status";

export type NavGate = "requires_verification";

export type NavItem = {
  label: string;
  href: Route;
  icon: LucideIcon;
  gate?: NavGate | undefined;
};

export type NavSection = {
  label: string;
  items: NavItem[];
};

const overview: NavSection = {
  label: "Overview",
  items: [{ label: "Dashboard", href: "/dashboard", icon: LayoutDashboard }],
};

const notifications: NavSection = {
  label: "Activity",
  items: [{ label: "Notifications", href: "/notifications", icon: Bell }],
};

const producerNavigation: NavSection[] = [
  overview,
  {
    label: "Provenance",
    items: [
      { label: "Batches", href: "/batches", icon: Boxes },
      { label: "Facilities", href: "/facilities", icon: Factory },
    ],
  },
  {
    label: "Sustainability",
    items: [
      {
        label: "Claims",
        href: "/claims",
        icon: FileCheck2,
        gate: "requires_verification",
      },
      { label: "Credits", href: "/credits", icon: Leaf },
    ],
  },
  {
    label: "Trade",
    items: [
      { label: "Marketplace", href: "/marketplace", icon: Store },
      {
        label: "My listings",
        href: "/marketplace/my-listings",
        icon: ListChecks,
        gate: "requires_verification",
      },
      {
        label: "My retirements",
        href: "/marketplace/my-retirements",
        icon: ShoppingCart,
      },
    ],
  },
  {
    label: "Organization",
    items: [
      { label: "Team", href: "/organization/users", icon: Users },
      { label: "API keys", href: "/organization/api-keys", icon: KeyRound },
      { label: "Wallet", href: "/organization/wallet", icon: Wallet },
      { label: "Settings", href: "/organization/settings", icon: Settings },
      { label: "Billing", href: "/billing", icon: CreditCard },
    ],
  },
  notifications,
];

const logisticsNavigation: NavSection[] = [
  overview,
  {
    label: "Provenance",
    items: [{ label: "Batches", href: "/batches", icon: Boxes }],
  },
  {
    label: "Organization",
    items: [
      { label: "Team", href: "/organization/users", icon: Users },
      { label: "API keys", href: "/organization/api-keys", icon: KeyRound },
      { label: "Settings", href: "/organization/settings", icon: Settings },
      { label: "Billing", href: "/billing", icon: CreditCard },
    ],
  },
  notifications,
];

const buyerNavigation: NavSection[] = [
  overview,
  {
    label: "Trade",
    items: [
      { label: "Marketplace", href: "/marketplace", icon: Store },
      {
        label: "My retirements",
        href: "/marketplace/my-retirements",
        icon: ShoppingCart,
      },
    ],
  },
  {
    label: "Portfolio",
    items: [{ label: "Credits", href: "/credits", icon: Leaf }],
  },
  {
    label: "Organization",
    items: [
      { label: "Team", href: "/organization/users", icon: Users },
      { label: "Wallet", href: "/organization/wallet", icon: Wallet },
      { label: "Settings", href: "/organization/settings", icon: Settings },
    ],
  },
  notifications,
];

export const verifierNavigation: NavSection[] = [
  {
    label: "Review",
    items: [
      { label: "Queue", href: "/verifier/queue", icon: ClipboardCheck },
      { label: "History", href: "/verifier/history", icon: History },
    ],
  },
  notifications,
];

const navigationByOrganizationType: Record<OrganizationType, NavSection[]> = {
  manufacturer: producerNavigation,
  assembler: producerNavigation,
  logistics: logisticsNavigation,
  credit_buyer: buyerNavigation,
};

export const navigationFor = (organizationType: OrganizationType) =>
  navigationByOrganizationType[organizationType];

export type GateVerdict =
  { allowed: true } | { allowed: false; reason: string };

export const evaluateGate = (
  gate: NavGate | undefined,
  verificationStatus: VerificationStatus,
  state: OrganizationState,
): GateVerdict => {
  if (!gate) {
    return { allowed: true };
  }

  if (state === "restricted") {
    return {
      allowed: false,
      reason:
        "Your organization is restricted following a fraud escalation. You can still view and export your data.",
    };
  }

  if (state === "read_only") {
    return {
      allowed: false,
      reason:
        "Your subscription lapsed past the grace period. Update your payment method to restore this.",
    };
  }

  if (gate === "requires_verification" && verificationStatus !== "verified") {
    return {
      allowed: false,
      reason:
        "Your organization must be verified against the business registry before you can use this.",
    };
  }

  return { allowed: true };
};

export const gateFor = (item: NavItem, organization: Organization) =>
  evaluateGate(item.gate, organization.verificationStatus, organization.state);

export const isActiveRoute = (pathname: string, href: string) =>
  pathname === href || pathname.startsWith(`${href}/`);
