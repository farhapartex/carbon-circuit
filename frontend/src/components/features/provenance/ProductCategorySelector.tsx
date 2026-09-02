"use client";

import { TriangleAlert } from "lucide-react";
import { productCategoryLabels } from "@/lib/labels";
import type { ProductCategory } from "@/lib/types";
import { cn } from "@/lib/utils";

type ProductCategorySelectorProps = {
  available: ProductCategory[];
  value: ProductCategory;
  onValueChange: (category: ProductCategory) => void;
  acknowledged: boolean;
  onAcknowledgedChange: (acknowledged: boolean) => void;
};

export function ProductCategorySelector({
  available,
  value,
  onValueChange,
  acknowledged,
  onAcknowledgedChange,
}: ProductCategorySelectorProps) {
  return (
    <div className="space-y-5">
      <fieldset className="space-y-3">
        <legend className="font-medium">Product category</legend>
        <div className="grid gap-3 sm:grid-cols-2">
          {available.map((category) => (
            <label
              key={category}
              className={cn(
                "flex cursor-pointer items-start gap-3 rounded-lg border px-4 py-3",
                value === category
                  ? "border-primary-600 bg-primary-50"
                  : "hover:border-neutral-300 border-neutral-200",
              )}
            >
              <input
                type="radio"
                name="productCategory"
                value={category}
                checked={value === category}
                onChange={() => onValueChange(category)}
                className="mt-1"
              />
              <span>
                <span className="block font-medium">
                  {productCategoryLabels[category]}
                </span>
                <span className="block text-caption text-neutral-600">
                  {category === "electronics"
                    ? "Component type and lot number, with a five-stage checkpoint sequence"
                    : "Attributes and checkpoint expectations for this category"}
                </span>
              </span>
            </label>
          ))}
        </div>
      </fieldset>

      <div className="rounded-md border border-warning-600 bg-warning-50 px-4 py-3">
        <p className="flex items-center gap-2 font-medium text-warning-700">
          <TriangleAlert className="size-4 shrink-0" aria-hidden />
          This choice is permanent
        </p>
        <p className="mt-1 text-caption text-pretty text-warning-700">
          A batch&apos;s product category cannot be changed after it is created.
          It fixes the expected checkpoint sequence its Provenance Score is
          measured against, and which claim types its facility can file.
          Creating a new batch is the only way to correct it.
        </p>
        <label className="mt-3 flex items-start gap-2 text-caption text-warning-700">
          <input
            type="checkbox"
            checked={acknowledged}
            onChange={(event) => onAcknowledgedChange(event.target.checked)}
            className="mt-0.5"
          />
          I understand this cannot be changed later.
        </label>
      </div>
    </div>
  );
}
