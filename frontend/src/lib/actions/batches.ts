"use server";

import { revalidatePath } from "next/cache";
import {
  createBatch,
  fetchBatches,
  logCheckpoint,
  type BatchDetail,
  type BatchDraft,
  type BatchPage,
  type CheckpointDraft,
} from "@/lib/api/batches";
import { GatewayError } from "@/lib/api/gateway";
import { auth0 } from "@/lib/auth0";

export type BatchResult =
  { ok: true; detail: BatchDetail } | { ok: false; code: string };

export type CheckpointResult = { ok: true } | { ok: false; code: string };

const failed = (error: unknown): { ok: false; code: string } => {
  if (error instanceof GatewayError) {
    return { ok: false, code: error.code };
  }
  throw error;
};

export const loadMoreBatches = async (after: string): Promise<BatchPage> => {
  const { token } = await auth0.getAccessToken();
  return fetchBatches(token, after);
};

export const submitBatch = async (
  draft: BatchDraft,
  idempotencyKey: string,
): Promise<BatchResult> => {
  try {
    const { token } = await auth0.getAccessToken();
    const detail = await createBatch(token, draft, idempotencyKey);

    revalidatePath("/batches");

    return { ok: true, detail };
  } catch (error) {
    return failed(error);
  }
};

export const submitCheckpoint = async (
  batchId: string,
  draft: CheckpointDraft,
  idempotencyKey: string,
): Promise<CheckpointResult> => {
  try {
    const { token } = await auth0.getAccessToken();
    await logCheckpoint(token, batchId, draft, idempotencyKey);

    revalidatePath(`/batches/${batchId}`);
    revalidatePath("/batches");

    return { ok: true };
  } catch (error) {
    return failed(error);
  }
};
