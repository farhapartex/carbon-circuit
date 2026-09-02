"use server";

import { listBatches } from "@/lib/fixtures";
import type { Batch, CursorMeta } from "@/lib/types";

export type BatchPage = {
  items: Batch[];
  meta: CursorMeta;
};

export const loadMoreBatches = async (after: string): Promise<BatchPage> => {
  const page = await listBatches(after);
  return { items: page.items, meta: page.meta };
};
