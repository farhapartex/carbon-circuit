import { z } from "zod";
import type { ProductCategory } from "@/lib/types";

const PRODUCT_CATEGORIES: [ProductCategory, ...ProductCategory[]] = [
  "electronics",
  "agriculture",
  "pharma",
  "textiles",
];

const PUBLIC_REFERENCE = /^[0-9A-Za-z]{22}$/;

export const batchDraftSchema = z.object({
  productCategory: z.enum(PRODUCT_CATEGORIES),
  acknowledgedCategory: z.boolean().refine((value) => value, {
    message: "Confirm you understand the category cannot be changed.",
  }),
  originatingFacilityId: z.string().min(1, "Select the originating facility."),
  componentType: z
    .string()
    .min(2, "Describe what this batch contains.")
    .max(160),
  lotNumber: z.string().max(64).optional(),
  externalId: z.string().max(128).optional(),
  quantity: z
    .string()
    .regex(
      /^\d+(\.\d{1,6})?$/,
      "Quantity must be a plain number with up to six decimal places.",
    )
    .refine((value) => Number(value) > 0, {
      message: "Quantity must be greater than zero.",
    }),
  unit: z.string().min(1, "Name the unit this quantity is counted in.").max(32),
  producedAt: z
    .string()
    .min(1, "When was this batch produced?")
    .refine((value) => value <= new Date().toISOString().slice(0, 10), {
      message: "A batch cannot be produced in the future.",
    }),
  parentReferences: z.array(
    z.object({
      value: z
        .string()
        .regex(
          PUBLIC_REFERENCE,
          "A public batch reference is 22 letters and digits.",
        ),
    }),
  ),
});

export type BatchDraftValues = z.infer<typeof batchDraftSchema>;

export const BATCH_STEPS = ["category", "details", "review"] as const;

export type BatchStep = (typeof BATCH_STEPS)[number];

export const stepFields: Record<BatchStep, (keyof BatchDraftValues)[]> = {
  category: ["productCategory", "acknowledgedCategory"],
  details: [
    "originatingFacilityId",
    "componentType",
    "lotNumber",
    "externalId",
    "quantity",
    "unit",
    "producedAt",
    "parentReferences",
  ],
  review: [],
};
