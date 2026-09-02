import type { Metadata } from "next";
import { InviteMemberDialog } from "@/components/features/settings/InviteMemberDialog";
import { MemberList } from "@/components/features/settings/MemberList";
import { PendingInvitationList } from "@/components/features/settings/PendingInvitationList";
import { fetchTeam } from "@/lib/api/members";
import { auth0 } from "@/lib/auth0";
import { tenancyOf } from "@/lib/session";
import type { OrganizationInvitation, OrganizationUser } from "@/lib/types";

export const metadata: Metadata = { title: "Members" };

export default async function SettingsMembersPage() {
  const session = await auth0.getSession();
  const { token } = await auth0.getAccessToken();
  const team = await fetchTeam(token);
  const tenancy = tenancyOf(session);

  const members: OrganizationUser[] = team.members.map((member) => ({
    id: member.userId,
    name: member.name,
    email: member.email,
    role: member.role,
    mfaEnabled: member.mfaEnrolled,
    lastActiveAt: member.lastActiveAt,
    invitedAt: member.joinedAt ?? "",
  }));

  const invitations: OrganizationInvitation[] = team.invitations.map(
    (invitation) => ({
      id: invitation.id,
      email: invitation.email,
      role: invitation.role,
      state: invitation.state,
      invitedByName: invitation.invitedByName,
      invitedAt: invitation.invitedAt,
      expiresAt: invitation.expiresAt,
    }),
  );

  const signedInMember = team.members.find(
    (member) => member.email === session?.user.email,
  );

  return (
    <>
      <div className="flex flex-wrap items-start justify-between gap-4">
        <p className="max-w-xl text-caption text-pretty text-neutral-600">
          Everyone here can act on behalf of this organization. Owners and
          administrators must have two-factor authentication enabled, because
          those roles can move the organization&apos;s assets.
        </p>
        <InviteMemberDialog canGrantOwner={tenancy?.role === "owner"} />
      </div>

      <MemberList
        members={members}
        signedInUserId={signedInMember?.userId ?? ""}
      />
      <PendingInvitationList invitations={invitations} />
    </>
  );
}
