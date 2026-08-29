"use client";

import { useState } from "react";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { MFARequiredBadge } from "@/components/shared/StatusBadges";
import { StatusPill } from "@/components/shared/StatusPill";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { OrganizationRole, OrganizationUser } from "@/lib/types";
import { useToastQueueStore } from "@/stores/toast-queue";

const ROLE_PRESENTATION: Record<
  OrganizationRole,
  { label: string; variant: "primary" | "info" | "neutral" }
> = {
  owner: { label: "Owner", variant: "primary" },
  admin: { label: "Administrator", variant: "info" },
  member: { label: "Member", variant: "neutral" },
};

const ROLES_REQUIRING_MFA: OrganizationRole[] = ["owner", "admin"];

type MemberListProps = {
  members: OrganizationUser[];
  signedInUserId: string;
};

export function MemberList({ members, signedInUserId }: MemberListProps) {
  const [pendingRevoke, setPendingRevoke] = useState<OrganizationUser | null>(
    null,
  );
  const pushToast = useToastQueueStore((state) => state.pushToast);
  const ownerCount = members.filter((member) => member.role === "owner").length;

  const revoke = () => {
    if (!pendingRevoke) return;
    pushToast({
      tone: "success",
      title: "Access revoked",
      description: `${pendingRevoke.name} can no longer act for this organization.`,
    });
    setPendingRevoke(null);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Members</CardTitle>
      </CardHeader>
      <CardContent>
        <ul className="divide-y divide-neutral-200">
          {members.map((member) => {
            const isSelf = member.id === signedInUserId;
            const isLastOwner = member.role === "owner" && ownerCount === 1;
            const mfaRequired = ROLES_REQUIRING_MFA.includes(member.role);

            return (
              <li
                key={member.id}
                className="flex flex-wrap items-center gap-3 py-4 first:pt-0 last:pb-0"
              >
                <span className="min-w-0 flex-1">
                  <span className="flex flex-wrap items-center gap-2">
                    <span className="font-medium">{member.name}</span>
                    <StatusPill
                      presentation={ROLE_PRESENTATION[member.role]}
                      showDot={false}
                    />
                    {isSelf ? (
                      <span className="text-caption text-neutral-600">You</span>
                    ) : null}
                  </span>
                  <span className="block text-caption text-neutral-600">
                    {member.email}
                    {member.lastActiveAt ? (
                      <>
                        {" · last active "}
                        <TimestampDisplay value={member.lastActiveAt} />
                      </>
                    ) : (
                      " · has not signed in yet"
                    )}
                  </span>
                </span>

                {mfaRequired && !member.mfaEnabled ? (
                  <StatusPill
                    presentation={{ label: "MFA missing", variant: "danger" }}
                  />
                ) : mfaRequired ? (
                  <MFARequiredBadge />
                ) : null}

                <Button
                  variant="outline"
                  size="sm"
                  disabled={isLastOwner}
                  title={
                    isLastOwner
                      ? "An organization must always have at least one Owner"
                      : undefined
                  }
                  onClick={() => setPendingRevoke(member)}
                >
                  Revoke access
                </Button>
              </li>
            );
          })}
        </ul>
      </CardContent>

      <ConfirmDialog
        open={pendingRevoke !== null}
        onOpenChange={(open) => {
          if (!open) setPendingRevoke(null);
        }}
        title={`Revoke access for ${pendingRevoke?.name ?? ""}?`}
        description="They lose access immediately, not when their session expires."
        consequence="Any batches they created and checkpoints they logged stay on the record, because removing them would break another organization's provenance chain."
        confirmLabel="Revoke access"
        destructive
        onConfirm={revoke}
      />
    </Card>
  );
}
