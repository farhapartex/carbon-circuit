import type { Route } from "next";
import { redirect } from "next/navigation";

export default function LoginPage() {
  redirect("/auth/login?returnTo=%2Fdashboard" as Route);
}
