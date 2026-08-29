import { MarketingShell } from "@/components/layout/MarketingShell";

export default function MarketingRouteLayout({ children }: LayoutProps<"/">) {
  return <MarketingShell>{children}</MarketingShell>;
}
