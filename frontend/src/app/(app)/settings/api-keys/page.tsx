import type { Metadata } from "next";
import Link from "next/link";
import { CopyButton } from "@/components/shared/CopyButton";
import { EmptyState } from "@/components/shared/EmptyState";
import { StatusPill } from "@/components/shared/StatusPill";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { fetchPlans } from "@/lib/api/plans";
import { getSubscription, listApiKeys } from "@/lib/fixtures";

export const metadata: Metadata = { title: "API keys" };

export default async function SettingsApiKeysPage() {
  const subscription = await getSubscription();
  const plans = await fetchPlans();
  const plan = plans.find(
    (candidate) => candidate.tier === subscription.planTier,
  );
  const keys = await listApiKeys();

  if (!plan?.apiKeyLimit) {
    return (
      <EmptyState
        title="API access is not included in your plan"
        description="Growth and Enterprise organizations can submit batches and checkpoints straight from their own systems. Your current plan uses portal entry only."
        action={
          <Button asChild>
            <Link href="/settings/billing/plans">Compare plans</Link>
          </Button>
        }
      />
    );
  }

  const activeKeys = keys.filter((key) => key.revokedAt === null);

  return (
    <>
      <div className="flex flex-wrap items-start justify-between gap-4">
        <p className="max-w-xl text-caption text-pretty text-neutral-600">
          Keys let your ERP or warehouse system submit batches and checkpoints
          directly. Every submission carries your own external identifier, so
          replaying a day after an outage creates no duplicates. Using{" "}
          {activeKeys.length} of {plan.apiKeyLimit} active keys.
        </p>
        <Button disabled={activeKeys.length >= plan.apiKeyLimit}>
          Create API key
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Keys</CardTitle>
        </CardHeader>
        <CardContent>
          <ul className="divide-y divide-neutral-200">
            {keys.map((key) => (
              <li
                key={key.id}
                className="flex flex-wrap items-center gap-3 py-4 first:pt-0 last:pb-0"
              >
                <span className="min-w-0 flex-1">
                  <span className="flex flex-wrap items-center gap-2">
                    <span className="font-medium">{key.name}</span>
                    <StatusPill
                      presentation={
                        key.revokedAt
                          ? { label: "Revoked", variant: "neutral" }
                          : { label: "Active", variant: "success" }
                      }
                    />
                  </span>
                  <span className="flex flex-wrap items-center gap-2 text-caption text-neutral-600">
                    <span className="font-mono">{key.prefix}…</span>
                    <CopyButton value={key.prefix} label="key prefix" />
                    {key.lastUsedAt ? (
                      <>
                        {" · last used "}
                        <TimestampDisplay value={key.lastUsedAt} />
                      </>
                    ) : (
                      " · never used"
                    )}
                  </span>
                </span>
                {key.revokedAt ? null : (
                  <Button variant="outline" size="sm">
                    Revoke
                  </Button>
                )}
              </li>
            ))}
          </ul>
        </CardContent>
      </Card>
    </>
  );
}
