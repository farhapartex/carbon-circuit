"use client";

import { useState, useTransition } from "react";
import { StatusPill } from "@/components/shared/StatusPill";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { revokeSessionAction } from "@/lib/actions/sessions";
import type { SessionRecord } from "@/lib/api/sessions";

const failureMessage = (code: string) => {
  if (code === "RESOURCE_NOT_FOUND") {
    return "That session has already ended.";
  }
  return "We could not sign that device out. Please try again.";
};

export function ActiveSessionList({ sessions }: { sessions: SessionRecord[] }) {
  const [failure, setFailure] = useState<string | null>(null);
  const [revoking, setRevoking] = useState<string | null>(null);
  const [pending, startTransition] = useTransition();

  const revoke = (sessionId: string) => {
    setFailure(null);
    setRevoking(sessionId);

    startTransition(async () => {
      try {
        const result = await revokeSessionAction(
          sessionId,
          crypto.randomUUID(),
        );
        if (!result.ok) setFailure(failureMessage(result.code));
      } catch (error) {
        setFailure(
          `The request could not be sent: ${error instanceof Error ? error.message : "unknown error"}`,
        );
      } finally {
        setRevoking(null);
      }
    });
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Active sessions</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {failure ? (
          <div
            role="alert"
            className="rounded-md border border-danger-600 bg-danger-50 px-4 py-3"
          >
            <p className="text-helper text-danger-700">{failure}</p>
          </div>
        ) : null}

        {sessions.length === 0 ? (
          <p className="text-caption text-pretty text-neutral-600">
            No sessions recorded yet. This device will appear here shortly after
            your next request.
          </p>
        ) : (
          <ul className="divide-y divide-neutral-200">
            {sessions.map((session) => (
              <li
                key={session.id}
                className="flex flex-wrap items-center gap-3 py-3 first:pt-0 last:pb-0"
              >
                <span className="min-w-0">
                  <span className="flex flex-wrap items-center gap-2 font-medium">
                    {session.userAgent || "Unrecognised device"}
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
                  <Button
                    variant="outline"
                    size="sm"
                    className="ml-auto"
                    disabled={pending && revoking === session.id}
                    onClick={() => revoke(session.id)}
                  >
                    Sign out
                  </Button>
                )}
              </li>
            ))}
          </ul>
        )}

        <p className="text-caption text-pretty text-neutral-600">
          Signing a device out revokes only that session. Its next request is
          refused even if its token has not expired.
        </p>
      </CardContent>
    </Card>
  );
}
