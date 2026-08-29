import type { Metadata } from "next";
import { InviteMemberDialog } from "@/components/features/settings/InviteMemberDialog";
import { MemberList } from "@/components/features/settings/MemberList";
import { PendingInvitationList } from "@/components/features/settings/PendingInvitationList";
import {
  getSignedInUser,
  listInvitations,
  listOrganizationUsers,
} from "@/lib/fixtures";

export const metadata: Metadata = { title: "Members" };

export default async function SettingsMembersPage() {
  const members = await listOrganizationUsers();
  const invitations = await listInvitations();
  const signedInUser = await getSignedInUser();

  return (
    <>
      <div className="flex flex-wrap items-start justify-between gap-4">
        <p className="max-w-xl text-caption text-pretty text-neutral-600">
          Everyone here can act on behalf of this organization. Owners and
          administrators must have two-factor authentication enabled, because
          those roles can move the organization&apos;s assets.
        </p>
        <InviteMemberDialog />
      </div>

      <MemberList members={members} signedInUserId={signedInUser.id} />
      <PendingInvitationList invitations={invitations} />
    </>
  );
}
