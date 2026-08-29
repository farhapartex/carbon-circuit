import type { Metadata } from "next";
import { CreditCard } from "lucide-react";
import Link from "next/link";
import { PlanUsageWidget } from "@/components/features/billing/PlanUsageWidget";
import { CancelSubscriptionDialog } from "@/components/features/settings/CancelSubscriptionDialog";
import { StatusPill } from "@/components/shared/StatusPill";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  getPaymentMethod,
  getPlanUsage,
  getSubscription,
  listInvoices,
  listPlans,
} from "@/lib/fixtures";
import type { SubscriptionState } from "@/lib/types";

export const metadata: Metadata = { title: "Billing" };

const SUBSCRIPTION_PRESENTATION: Record<
  SubscriptionState,
  { label: string; variant: "success" | "warning" | "danger" | "neutral" }
> = {
  active: { label: "Active", variant: "success" },
  grace_period: { label: "Payment failed", variant: "warning" },
  read_only: { label: "Read-only", variant: "danger" },
  cancelled: { label: "Cancelled", variant: "neutral" },
};

export default async function SettingsBillingPage() {
  const subscription = await getSubscription();
  const usage = await getPlanUsage();
  const paymentMethod = await getPaymentMethod();
  const invoices = await listInvoices();
  const plans = await listPlans();
  const plan = plans.find(
    (candidate) => candidate.tier === subscription.planTier,
  );

  return (
    <>
      <Card>
        <CardHeader className="flex-row items-start justify-between gap-4">
          <div className="space-y-1">
            <CardTitle>{plan?.name ?? "Plan"}</CardTitle>
            <p className="text-caption text-neutral-600">
              {plan?.monthlyPriceUsd === "0"
                ? "Free. Buying and retiring credits costs nothing."
                : `$${plan?.monthlyPriceUsd} per month`}
              {subscription.renewsAt ? (
                <>
                  {" · renews "}
                  <TimestampDisplay value={subscription.renewsAt} dateOnly />
                </>
              ) : null}
            </p>
          </div>
          <StatusPill
            presentation={SUBSCRIPTION_PRESENTATION[subscription.state]}
          />
        </CardHeader>
        <CardContent className="flex flex-wrap gap-2">
          <Button asChild variant="outline" size="sm">
            <Link href="/settings/billing/plans">Change plan</Link>
          </Button>
          <CancelSubscriptionDialog planName={plan?.name ?? "your plan"} />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Usage this period</CardTitle>
          <p className="text-caption text-neutral-600">
            <TimestampDisplay value={usage.periodStart} dateOnly /> to{" "}
            <TimestampDisplay value={usage.periodEnd} dateOnly />
          </p>
        </CardHeader>
        <CardContent>
          <PlanUsageWidget usage={usage} />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Payment method</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-wrap items-center gap-3">
          <CreditCard className="size-4 text-neutral-600" aria-hidden />
          <span className="font-medium tabular-nums">
            {paymentMethod.brand} ending {paymentMethod.last4}
          </span>
          <span className="text-caption text-neutral-600 tabular-nums">
            expires {String(paymentMethod.expiryMonth).padStart(2, "0")}/
            {paymentMethod.expiryYear}
          </span>
          <Button variant="outline" size="sm" className="ml-auto">
            Update
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Invoices</CardTitle>
        </CardHeader>
        <CardContent>
          <ul className="divide-y divide-neutral-200">
            {invoices.map((invoice) => (
              <li
                key={invoice.id}
                className="flex flex-wrap items-center gap-3 py-3 first:pt-0 last:pb-0"
              >
                <span className="font-medium tabular-nums">
                  {invoice.number}
                </span>
                <span className="text-caption text-neutral-600">
                  <TimestampDisplay value={invoice.issuedAt} dateOnly />
                </span>
                <StatusPill
                  presentation={
                    invoice.status === "paid"
                      ? { label: "Paid", variant: "success" }
                      : invoice.status === "failed"
                        ? { label: "Failed", variant: "danger" }
                        : { label: "Open", variant: "warning" }
                  }
                />
                <span className="ml-auto font-medium tabular-nums">
                  ${invoice.amountUsd}
                </span>
              </li>
            ))}
          </ul>
        </CardContent>
      </Card>
    </>
  );
}
