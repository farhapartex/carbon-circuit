"use client";

import { Lock, PanelLeftClose, PanelLeftOpen } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  evaluateGate,
  isActiveRoute,
  navigationFor,
  verifierNavigation,
  type NavItem,
} from "@/lib/navigation";
import type { OrganizationState, OrganizationType } from "@/lib/types";
import type { VerificationStatus } from "@/lib/status";
import { useUiContextStore } from "@/stores/ui-context";
import { cn } from "@/lib/utils";

type SidebarProps = {
  organizationType: OrganizationType | null;
  organizationName: string;
  verificationStatus: VerificationStatus;
  organizationState: OrganizationState;
};

function NavEntry({
  item,
  active,
  collapsed,
  gateReason,
}: {
  item: NavItem;
  active: boolean;
  collapsed: boolean;
  gateReason: string | null;
}) {
  const Icon = item.icon;
  const shared = cn(
    "flex items-center gap-3 rounded-lg px-3 py-2 text-caption font-medium transition-colors",
    collapsed && "justify-center px-2",
  );

  if (gateReason) {
    return (
      <Tooltip>
        <TooltipTrigger asChild>
          <span
            aria-disabled="true"
            data-gated="true"
            className={cn(shared, "cursor-not-allowed text-neutral-400")}
          >
            <Icon className="size-4 shrink-0" aria-hidden />
            {collapsed ? null : (
              <>
                <span className="truncate">{item.label}</span>
                <Lock className="ml-auto size-3 shrink-0" aria-hidden />
              </>
            )}
          </span>
        </TooltipTrigger>
        <TooltipContent side="right" className="max-w-64">
          {gateReason}
        </TooltipContent>
      </Tooltip>
    );
  }

  return (
    <Link
      href={item.href}
      aria-current={active ? "page" : undefined}
      className={cn(
        shared,
        active
          ? "bg-primary-50 text-primary-800"
          : "text-neutral-600 hover:bg-neutral-100 hover:text-neutral-900",
      )}
    >
      <Icon className="size-4 shrink-0" aria-hidden />
      {collapsed ? <span className="sr-only">{item.label}</span> : item.label}
    </Link>
  );
}

export function Sidebar({
  organizationType,
  organizationName,
  verificationStatus,
  organizationState,
}: SidebarProps) {
  const sections =
    organizationType === null
      ? verifierNavigation
      : navigationFor(organizationType);
  const pathname = usePathname();
  const collapsed = useUiContextStore((state) => state.sidebarCollapsed);
  const toggleSidebar = useUiContextStore((state) => state.toggleSidebar);

  return (
    <TooltipProvider delayDuration={150}>
      <nav
        aria-label="Primary"
        data-collapsed={collapsed}
        className={cn(
          "flex shrink-0 flex-col gap-6 border-r border-neutral-200 bg-white py-4 transition-[width]",
          collapsed ? "w-16 px-2" : "w-64 px-3",
        )}
      >
        <div
          className={cn(
            "flex items-center gap-2",
            collapsed ? "justify-center" : "justify-between px-2",
          )}
        >
          {collapsed ? null : (
            <span className="truncate text-caption font-semibold">
              {organizationName}
            </span>
          )}
          <button
            type="button"
            onClick={toggleSidebar}
            aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
            aria-expanded={!collapsed}
            className="rounded-sm p-1 text-muted-foreground hover:text-foreground"
          >
            {collapsed ? (
              <PanelLeftOpen className="size-4" aria-hidden />
            ) : (
              <PanelLeftClose className="size-4" aria-hidden />
            )}
          </button>
        </div>

        {sections.map((section) => (
          <div key={section.label} className="space-y-1">
            {collapsed ? null : (
              <p className="px-3 text-caption font-medium tracking-wide text-neutral-400 uppercase">
                {section.label}
              </p>
            )}
            {section.items.map((item) => {
              const verdict = evaluateGate(
                item.gate,
                verificationStatus,
                organizationState,
              );
              return (
                <NavEntry
                  key={item.href}
                  item={item}
                  active={isActiveRoute(pathname, item.href)}
                  collapsed={collapsed}
                  gateReason={verdict.allowed ? null : verdict.reason}
                />
              );
            })}
          </div>
        ))}
      </nav>
    </TooltipProvider>
  );
}
