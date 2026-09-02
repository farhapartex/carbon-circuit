"use server";

import { revalidatePath } from "next/cache";
import { GatewayError } from "@/lib/api/gateway";
import {
  changeMemberRole,
  inviteMember,
  revokeInvitation,
  revokeMember,
  type IssuedInvitation,
} from "@/lib/api/members";
import { auth0 } from "@/lib/auth0";
import type { OrganizationRole } from "@/lib/types/organization";

export type InviteResult =
  { ok: true; issued: IssuedInvitation } | { ok: false; code: string };

export type MemberActionResult = { ok: true } | { ok: false; code: string };

const accessToken = async () => (await auth0.getAccessToken()).token;

const failed = (error: unknown): { ok: false; code: string } => {
  if (error instanceof GatewayError) {
    return { ok: false, code: error.code };
  }
  throw error;
};

export const submitInvitation = async (
  email: string,
  role: OrganizationRole,
  idempotencyKey: string,
): Promise<InviteResult> => {
  try {
    const issued = await inviteMember(
      await accessToken(),
      email,
      role,
      idempotencyKey,
    );
    revalidatePath("/settings/members");
    return { ok: true, issued };
  } catch (error) {
    return failed(error);
  }
};

export const submitRoleChange = async (
  userId: string,
  role: OrganizationRole,
  idempotencyKey: string,
): Promise<MemberActionResult> => {
  try {
    await changeMemberRole(await accessToken(), userId, role, idempotencyKey);
    revalidatePath("/settings/members");
    return { ok: true };
  } catch (error) {
    return failed(error);
  }
};

export const submitMemberRevocation = async (
  userId: string,
  idempotencyKey: string,
): Promise<MemberActionResult> => {
  try {
    await revokeMember(await accessToken(), userId, idempotencyKey);
    revalidatePath("/settings/members");
    return { ok: true };
  } catch (error) {
    return failed(error);
  }
};

export const submitInvitationRevocation = async (
  invitationId: string,
  idempotencyKey: string,
): Promise<MemberActionResult> => {
  try {
    await revokeInvitation(await accessToken(), invitationId, idempotencyKey);
    revalidatePath("/settings/members");
    return { ok: true };
  } catch (error) {
    return failed(error);
  }
};
