import type { Metadata } from "next";
import { MailCheck, MailWarning } from "lucide-react";
import { ActiveSessionList } from "@/components/features/settings/ActiveSessionList";
import { MfaSettingsCard } from "@/components/features/settings/MfaSettingsCard";
import { AddressDisplay } from "@/components/shared/AddressDisplay";
import { StatusPill } from "@/components/shared/StatusPill";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { fetchMe } from "@/lib/api/me";
import { fetchSessions } from "@/lib/api/sessions";
import { auth0 } from "@/lib/auth0";
import { getMfaSettings } from "@/lib/fixtures";

export const metadata: Metadata = { title: "Profile" };

const ROLE_LABELS = {
  owner: "Organization Owner",
  admin: "Administrator",
  member: "Member",
} as const;

export default async function SettingsProfilePage() {
  const { token } = await auth0.getAccessToken();

  const [session, sessions, mfa] = await Promise.all([
    fetchMe(token),
    fetchSessions(token),
    getMfaSettings(),
  ]);

  const { user, organization } = session;

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
                {organization ? (
                  <StatusPill
                    presentation={{
                      label: ROLE_LABELS[organization.role],
                      variant:
                        organization.role === "owner" ? "primary" : "neutral",
                    }}
                  />
                ) : (
                  <span className="text-caption text-neutral-600">
                    No organization yet
                  </span>
                )}
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
              <Button variant="outline" size="sm" disabled>
                Disconnect
              </Button>
            </div>
          ) : (
            <>
              <Button variant="outline" size="sm" disabled>
                Connect a personal wallet
              </Button>
              <p className="text-caption text-pretty text-neutral-600">
                Not available yet. Binding a personal wallet needs a signed
                proof of ownership, the same as the Treasury Address does, and
                that flow is not built.
              </p>
            </>
          )}
        </CardContent>
      </Card>

      <ActiveSessionList sessions={sessions} />
    </>
  );
}
