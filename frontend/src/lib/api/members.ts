import "server-only";
import {
  gatewayDelete,
  gatewayGet,
  gatewayPatch,
  gatewayPost,
} from "@/lib/api/gateway";
import type { OrganizationRole } from "@/lib/types/organization";

export type InvitationState = "pending" | "accepted" | "revoked" | "expired";

type ApiMember = {
  user_id: string;
  email: string;
  name: string;
  role: OrganizationRole;
  mfa_enrolled: boolean;
  joined_at: string | null;
  last_active_at: string | null;
};

type ApiInvitation = {
  id: string;
  email: string;
  role: OrganizationRole;
  state: InvitationState;
  invited_at: string;
  expires_at: string;
};

type ApiTeam = {
  members: ApiMember[];
  invitations: ApiInvitation[];
};

export type Member = {
  userId: string;
  email: string;
  name: string;
  role: OrganizationRole;
  mfaEnrolled: boolean;
  joinedAt: string | null;
  lastActiveAt: string | null;
};

export type Invitation = {
  id: string;
  email: string;
  role: OrganizationRole;
  state: InvitationState;
  invitedAt: string;
  expiresAt: string;
};

export type Team = {
  members: Member[];
  invitations: Invitation[];
};

export type IssuedInvitation = {
  invitation: Invitation;
  token: string;
};

const toMember = (member: ApiMember): Member => ({
  userId: member.user_id,
  email: member.email,
  name: member.name,
  role: member.role,
  mfaEnrolled: member.mfa_enrolled,
  joinedAt: member.joined_at,
  lastActiveAt: member.last_active_at,
});

const toInvitation = (invitation: ApiInvitation): Invitation => ({
  id: invitation.id,
  email: invitation.email,
  role: invitation.role,
  state: invitation.state,
  invitedAt: invitation.invited_at,
  expiresAt: invitation.expires_at,
});

export const fetchTeam = async (token: string): Promise<Team> => {
  const team = await gatewayGet<ApiTeam>("/v1/members", token);

  return {
    members: team.members.map(toMember),
    invitations: team.invitations.map(toInvitation),
  };
};

export const inviteMember = async (
  token: string,
  email: string,
  role: OrganizationRole,
  idempotencyKey: string,
): Promise<IssuedInvitation> => {
  const issued = await gatewayPost<{
    invitation: ApiInvitation;
    token: string;
  }>("/v1/invitations", token, { email, role }, idempotencyKey);

  return {
    invitation: toInvitation(issued.invitation),
    token: issued.token,
  };
};

export const changeMemberRole = async (
  token: string,
  userId: string,
  role: OrganizationRole,
  idempotencyKey: string,
): Promise<void> => {
  await gatewayPatch<unknown>(
    `/v1/members/${userId}`,
    token,
    { role },
    idempotencyKey,
  );
};

export const revokeMember = async (
  token: string,
  userId: string,
  idempotencyKey: string,
): Promise<void> => {
  await gatewayDelete(`/v1/members/${userId}`, token, idempotencyKey);
};

export const revokeInvitation = async (
  token: string,
  invitationId: string,
  idempotencyKey: string,
): Promise<void> => {
  await gatewayDelete(`/v1/invitations/${invitationId}`, token, idempotencyKey);
};

export const acceptInvitation = async (
  token: string,
  invitationToken: string,
  idempotencyKey: string,
): Promise<{
  organizationId: string;
  organizationName: string;
  role: OrganizationRole;
}> => {
  const accepted = await gatewayPost<{
    organization_id: string;
    organization_name: string;
    role: OrganizationRole;
  }>(
    "/v1/invitations/accept",
    token,
    { token: invitationToken },
    idempotencyKey,
  );

  return {
    organizationId: accepted.organization_id,
    organizationName: accepted.organization_name,
    role: accepted.role,
  };
};
