"use client";

import type { ClaimDraftValues } from "@/components/features/claims/claimDraft";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { FacilityRecord } from "@/lib/api/facilities";
import {
  activityTypeLabels,
  gridRegionLabels,
  recycledMaterialLabels,
} from "@/lib/labels";

const decimalFormat = new Intl.NumberFormat("en-US", {
  maximumFractionDigits: 6,
});

const Row = ({ label, value }: { label: string; value: string }) => (
  <div className="flex flex-wrap items-baseline justify-between gap-4">
    <dt className="text-caption text-neutral-600">{label}</dt>
    <dd className="font-medium">{value}</dd>
  </div>
);

const figureRows = (values: ClaimDraftValues): [string, string][] => {
  if (values.activityType === "renewable_energy") {
    return [
      [
        "Verified renewable energy",
        `${decimalFormat.format(Number(values.verifiedKwh ?? 0))} kWh`,
      ],
      [
        "Grid region",
        values.gridRegion ? gridRegionLabels[values.gridRegion] : "Not set",
      ],
    ];
  }

  if (values.activityType === "reduced_emission_logistics") {
    return [
      [
        "Tonne-kilometres",
        decimalFormat.format(Number(values.tonneKilometres ?? 0)),
      ],
      [
        "Actual factor",
        `${values.actualFactorKgPerTonneKm ?? "0"} kgCO2e/tonne-km`,
      ],
    ];
  }

  return [
    [
      "Material",
      values.material ? recycledMaterialLabels[values.material] : "Not set",
    ],
    [
      "Verified quantity",
      `${decimalFormat.format(Number(values.verifiedQuantity ?? 0))} ${values.quantityUnit ?? "tonne"}`,
    ],
  ];
};

type ClaimReviewStepProps = {
  values: ClaimDraftValues;
  facilities: FacilityRecord[];
};

export function ClaimReviewStep({ values, facilities }: ClaimReviewStepProps) {
  const facility = facilities.find(
    (candidate) => candidate.id === values.facilityId,
  );

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Review this claim</CardTitle>
        </CardHeader>
        <CardContent>
          <dl className="space-y-3">
            <Row label="Facility" value={facility?.name ?? "Not selected"} />
            <Row
              label="Activity"
              value={activityTypeLabels[values.activityType]}
            />
            <Row label="Vintage" value={values.vintageYear} />
            <Row
              label="Claim period"
              value={`${values.periodStart} — ${values.periodEnd}`}
            />
            {figureRows(values).map(([label, value]) => (
              <Row key={label} label={label} value={value} />
            ))}
            <Row
              label="Requested amount"
              value={`${decimalFormat.format(Number(values.requestedAmount))} tCO2e`}
            />
            <Row
              label="Exclusivity attested"
              value={values.exclusivityAttested ? "Yes" : "Not yet"}
            />
          </dl>
        </CardContent>
      </Card>

      {facility ? (
        <p className="text-caption text-pretty text-neutral-600">
          This facility carries a ceiling discount of{" "}
          {facility.ceilingDiscountFactor}, so its ceiling is computed from that
          share of its capacity. Your requested amount cannot be issued above
          the ceiling regardless of what a verifier approves.
        </p>
      ) : null}
    </div>
  );
}
