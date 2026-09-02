"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { productCategoryLabels } from "@/lib/labels";
import type { BatchDraftValues } from "@/components/features/provenance/batchDraft";
import type { FacilityRecord } from "@/lib/api/facilities";

const numberFormat = new Intl.NumberFormat("en-US");

const Row = ({ label, value }: { label: string; value: string }) => (
  <div className="flex flex-wrap items-baseline justify-between gap-4">
    <dt className="text-caption text-neutral-600">{label}</dt>
    <dd className="font-medium">{value}</dd>
  </div>
);

type ReviewSummaryStepProps = {
  values: BatchDraftValues;
  facilities: FacilityRecord[];
};

export function ReviewSummaryStep({
  values,
  facilities,
}: ReviewSummaryStepProps) {
  const facility = facilities.find(
    (candidate) => candidate.id === values.originatingFacilityId,
  );

  const declared = values.parentReferences.filter(
    (parent) => parent.value.length > 0,
  );

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Review this batch</CardTitle>
        </CardHeader>
        <CardContent>
          <dl className="space-y-3">
            <Row
              label="Product category"
              value={productCategoryLabels[values.productCategory]}
            />
            <Row
              label="Originating facility"
              value={facility?.name ?? "Not selected"}
            />
            <Row label="Component type" value={values.componentType} />
            {values.lotNumber ? (
              <Row label="Lot number" value={values.lotNumber} />
            ) : null}
            <Row
              label="Quantity"
              value={`${numberFormat.format(Number(values.quantity))} ${values.unit}`}
            />
            <Row label="Produced" value={values.producedAt} />
            {values.externalId ? (
              <Row label="Your reference" value={values.externalId} />
            ) : null}
            <Row
              label="Component batches declared"
              value={declared.length === 0 ? "None" : String(declared.length)}
            />
          </dl>
        </CardContent>
      </Card>

      {declared.length > 0 ? (
        <Card>
          <CardHeader>
            <CardTitle>Declared component batches</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-2">
              {declared.map((parent) => (
                <li key={parent.value} className="font-mono text-helper">
                  {parent.value}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      ) : null}

      <p className="text-caption text-pretty text-neutral-600">
        A batch cannot be deleted once created, and its product category is
        fixed. Its public reference is generated now and is what a consumer
        scanning the finished product will reach.
      </p>
    </div>
  );
}
