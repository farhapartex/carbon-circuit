import { AuthLayout } from "@/components/layout/AuthLayout";

export default function AuthRouteLayout({ children }: LayoutProps<"/">) {
  return <AuthLayout>{children}</AuthLayout>;
}
