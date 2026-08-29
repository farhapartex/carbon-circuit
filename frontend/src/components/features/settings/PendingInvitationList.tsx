"use client";

import { useState } from "react";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { StatusPill } from "@/components/shared/StatusPill";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { InvitationState, OrganizationInvitation } from "@/lib/types";
import { useToastQueueStore } from "@/stores/toast-queue";

const STATE_PRESENTATION: Record<
  InvitationState,
  { label: string; variant: "warning" | "success" | "neutral" }
> = {
  pending: { label: "Pending", variant: "warning" },
  accepted: { label: "Accepted", variant: "success" },
  revoked: { label: "Revoked", variant: "neutral" },
  expired: { label: "Expired", variant: "neutral" },
};

export function PendingInvitationList({
  invitations,
}: {
  invitations: OrganizationInvitation[];
}) {
  const [pendingRevoke, setPendingRevoke] =
    useState<OrganizationInvitation | null>(null);
  const pushToast = useToastQueueStore((state) => state.pushToast);

  if (invitations.length === 0) return null;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Invitations</CardTitle>
      </CardHeader>
      <CardContent>
        <ul className="divide-y divide-neutral-200">
          {invitations.map((invitation) => (
            <li
              key={invitation.id}
              className="flex flex-wrap items-center gap-3 py-4 first:pt-0 last:pb-0"
            >
              <span className="min-w-0 flex-1">
                <span className="flex flex-wrap items-center gap-2">
                  <span className="font-medium">{invitation.email}</span>
                  <StatusPill
                    presentation={STATE_PRESENTATION[invitation.state]}
                  />
                </span>
                <span className="block text-caption text-neutral-600">
                  Invited as {invitation.role} by {invitation.invitedByName} ·{" "}
                  {invitation.state === "pending" ? "expires " : "expired "}
                  <TimestampDisplay value={invitation.expiresAt} dateOnly />
                </span>
              </span>

              {invitation.state === "pending" ? (
                <>
                  <Button variant="ghost" size="sm">
                    Resend
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPendingRevoke(invitation)}
                  >
                    Revoke
                  </Button>
                </>
              ) : invitation.state === "expired" ? (
                <Button variant="outline" size="sm">
                  Invite again
                </Button>
              ) : null}
            </li>
          ))}
        </ul>
      </CardContent>

      <ConfirmDialog
        open={pendingRevoke !== null}
        onOpenChange={(open) => {
          if (!open) setPendingRevoke(null);
        }}
        title="Revoke this invitation?"
        description={`The link sent to ${pendingRevoke?.email ?? ""} stops working immediately.`}
        confirmLabel="Revoke invitation"
        destructive
        onConfirm={() => {
          pushToast({ tone: "success", title: "Invitation revoked" });
          setPendingRevoke(null);
        }}
      />
    </Card>
  );
}
