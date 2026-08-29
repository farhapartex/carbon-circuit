import type { Metadata } from "next";
import { MailCheck, MailWarning } from "lucide-react";
import { MfaSettingsCard } from "@/components/features/settings/MfaSettingsCard";
import { AddressDisplay } from "@/components/shared/AddressDisplay";
import { StatusPill } from "@/components/shared/StatusPill";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  getMfaSettings,
  getSignedInUser,
  listActiveSessions,
} from "@/lib/fixtures";

export const metadata: Metadata = { title: "Profile" };

const ROLE_LABELS = {
  owner: "Organization Owner",
  admin: "Administrator",
  member: "Member",
} as const;

export default async function SettingsProfilePage() {
  const user = await getSignedInUser();
  const mfa = await getMfaSettings();
  const sessions = await listActiveSessions();

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Your profile</CardTitle>
        </CardHeader>
        <CardContent>
          <dl className="grid gap-4 sm:grid-cols-2">
            <div>
              <dt className="text-caption text-neutral-600">Name</dt>
              <dd className="font-medium">{user.name}</dd>
            </div>
            <div>
              <dt className="text-caption text-neutral-600">Email</dt>
              <dd className="flex flex-wrap items-center gap-2 font-medium">
                {user.email}
                {user.emailVerified ? (
                  <span className="inline-flex items-center gap-1 text-caption font-normal text-success-700">
                    <MailCheck className="size-3.5" aria-hidden />
                    Verified
                  </span>
                ) : (
                  <span className="inline-flex items-center gap-1 text-caption font-normal text-warning-700">
                    <MailWarning className="size-3.5" aria-hidden />
                    Unverified
                  </span>
                )}
              </dd>
            </div>
            <div>
              <dt className="text-caption text-neutral-600">Role</dt>
              <dd>
                {user.role ? (
                  <StatusPill
                    presentation={{
                      label: ROLE_LABELS[user.role],
                      variant: user.role === "owner" ? "primary" : "neutral",
                    }}
                  />
                ) : null}
              </dd>
            </div>
            <div>
              <dt className="text-caption text-neutral-600">Member since</dt>
              <dd className="font-medium">
                <TimestampDisplay value={user.createdAt} dateOnly />
              </dd>
            </div>
          </dl>
        </CardContent>
      </Card>

      <MfaSettingsCard settings={mfa} />

      <Card>
        <CardHeader>
          <CardTitle>Personal wallet</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-caption text-pretty text-neutral-600">
            A personal wallet signs actions on your organization&apos;s behalf
            where your role allows it. It never holds organization credits —
            those live at the Treasury Address, because an employee leaving
            should not take the company&apos;s assets with them.
          </p>
          {user.personalWalletAddress ? (
            <div className="flex flex-wrap items-center gap-3">
              <AddressDisplay address={user.personalWalletAddress} />
              <Button variant="outline" size="sm">
                Disconnect
              </Button>
            </div>
          ) : (
            <Button variant="outline" size="sm">
              Connect a personal wallet
            </Button>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Active sessions</CardTitle>
        </CardHeader>
        <CardContent>
          <ul className="divide-y divide-neutral-200">
            {sessions.map((session) => (
              <li
                key={session.id}
                className="flex flex-wrap items-center gap-3 py-3 first:pt-0 last:pb-0"
              >
                <span className="min-w-0">
                  <span className="flex flex-wrap items-center gap-2 font-medium">
                    {session.userAgent}
                    {session.current ? (
                      <StatusPill
                        presentation={{
                          label: "This device",
                          variant: "success",
                        }}
                      />
                    ) : null}
                  </span>
                  <span className="block text-caption text-neutral-600">
                    {session.ipAddress} · last seen{" "}
                    <TimestampDisplay value={session.lastSeenAt} />
                  </span>
                </span>
                {session.current ? null : (
                  <Button variant="outline" size="sm" className="ml-auto">
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
