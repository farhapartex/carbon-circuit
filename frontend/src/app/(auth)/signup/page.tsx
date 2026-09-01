import type { Route } from "next";
import { redirect } from "next/navigation";

export default function SignupPage() {
  redirect("/auth/login?screen_hint=signup&returnTo=%2Fdashboard" as Route);
}
