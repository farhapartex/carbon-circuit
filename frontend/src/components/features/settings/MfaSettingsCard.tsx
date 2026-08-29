import { KeyRound, ShieldAlert, ShieldCheck, Smartphone } from "lucide-react";
import { StatusPill } from "@/components/shared/StatusPill";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { MfaMethodKind, MfaSettings } from "@/lib/types";

const METHOD_ICONS: Record<MfaMethodKind, typeof Smartphone> = {
  authenticator_app: Smartphone,
  sms: Smartphone,
  recovery_codes: KeyRound,
};

export function MfaSettingsCard({ settings }: { settings: MfaSettings }) {
  return (
    <Card>
      <CardHeader className="flex-row items-start justify-between gap-4">
        <div className="space-y-1">
          <CardTitle className="flex items-center gap-2">
            {settings.enabled ? (
              <ShieldCheck className="size-4 text-success-600" aria-hidden />
            ) : (
              <ShieldAlert className="size-4 text-danger-600" aria-hidden />
            )}
            Two-factor authentication
          </CardTitle>
          <p className="text-caption text-neutral-600">
            {settings.requiredByRole
              ? "Required for your role, because it can move your organization's assets."
              : "Recommended for every account."}
          </p>
        </div>
        <StatusPill
          presentation={
            settings.enabled
              ? { label: "Enabled", variant: "success" }
              : { label: "Not enabled", variant: "danger" }
          }
        />
      </CardHeader>

      <CardContent className="space-y-4">
        {settings.requiredByRole && !settings.enabled ? (
          <p
            role="alert"
            className="rounded-lg bg-danger-50 px-4 py-3 text-caption text-danger-700"
          >
            Your role requires two-factor authentication. Until you enrol, you
            cannot approve claims, change the Treasury Address, or manage
            members.
          </p>
        ) : null}

        <ul className="divide-y divide-neutral-200">
          {settings.methods.map((method) => {
            const Icon = METHOD_ICONS[method.kind];
            return (
              <li
                key={method.kind}
                className="flex flex-wrap items-center gap-3 py-3 first:pt-0 last:pb-0"
              >
                <Icon
                  className="size-4 shrink-0 text-neutral-600"
                  aria-hidden
                />
                <span className="min-w-0">
                  <span className="flex flex-wrap items-center gap-2">
                    <span className="font-medium">{method.label}</span>
                    {method.isDefault ? (
                      <StatusPill
                        presentation={{ label: "Default", variant: "primary" }}
                        showDot={false}
                      />
                    ) : null}
                  </span>
                  {method.detail ? (
                    <span className="block text-caption text-neutral-600">
                      {method.detail}
                      {method.enrolledAt ? (
                        <>
                          {" · added "}
                          <TimestampDisplay
                            value={method.enrolledAt}
                            dateOnly
                          />
                        </>
                      ) : null}
                    </span>
                  ) : null}
                </span>
                <Button variant="outline" size="sm" className="ml-auto">
                  {method.kind === "recovery_codes" ? "Regenerate" : "Manage"}
                </Button>
              </li>
            );
          })}
        </ul>

        {settings.lastVerifiedAt ? (
          <p className="text-caption text-neutral-600">
            Last verified <TimestampDisplay value={settings.lastVerifiedAt} />.
          </p>
        ) : null}
      </CardContent>
    </Card>
  );
}
