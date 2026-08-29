import { PublicLayout } from "@/components/layout/PublicLayout";

export default function PublicRouteLayout({ children }: LayoutProps<"/">) {
  return <PublicLayout>{children}</PublicLayout>;
}
