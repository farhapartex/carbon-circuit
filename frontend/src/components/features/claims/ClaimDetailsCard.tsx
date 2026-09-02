import { CreditAmountDisplay } from "@/components/shared/CreditAmountDisplay";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { activityTypeLabels, shippingMethodLabels } from "@/lib/labels";
import type { ClaimFigures, SustainabilityClaim } from "@/lib/types";

const decimalFormat = new Intl.NumberFormat("en-US", {
  maximumFractionDigits: 6,
});

const figureRows = (figures: ClaimFigures): [string, string][] => {
  switch (figures.activityType) {
    case "renewable_energy":
      return [
        [
          "Verified renewable energy",
          `${decimalFormat.format(Number(figures.verifiedKwh))} kWh`,
        ],
        ["Grid region", figures.gridRegion],
      ];
    case "reduced_emission_logistics":
      return [
        [
          "Tonne-kilometres shipped",
          decimalFormat.format(Number(figures.tonneKilometres)),
        ],
        ["Shipping method", shippingMethodLabels[figures.shippingMethod]],
        [
          "Actual factor",
          `${figures.actualFactorKgPerTonneKm} kgCO2e/tonne-km`,
        ],
      ];
    case "responsible_sourcing":
      return [
        ["Material", figures.material],
        [
          "Verified quantity",
          `${decimalFormat.format(Number(figures.verifiedQuantity))} ${figures.quantityUnit}`,
        ],
      ];
  }
};

const Row = ({ label, value }: { label: string; value: React.ReactNode }) => (
  <div className="flex flex-wrap items-baseline justify-between gap-4">
    <dt className="text-caption text-neutral-600">{label}</dt>
    <dd className="font-medium">{value}</dd>
  </div>
);

export function ClaimDetailsCard({ claim }: { claim: SustainabilityClaim }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Claim details</CardTitle>
      </CardHeader>
      <CardContent>
        <dl className="space-y-3">
          <Row label="Facility" value={claim.facilityName} />
          <Row
            label="Activity"
            value={activityTypeLabels[claim.activityType]}
          />
          <Row label="Vintage" value={claim.vintageYear} />
          <Row
            label="Claim period"
            value={
              <>
                <TimestampDisplay value={claim.periodStart} dateOnly /> —{" "}
                <TimestampDisplay value={claim.periodEnd} dateOnly />
              </>
            }
          />
          {figureRows(claim.figures).map(([label, value]) => (
            <Row key={label} label={label} value={value} />
          ))}
          <Row
            label="Requested amount"
            value={<CreditAmountDisplay amount={claim.requestedAmount} />}
          />
          {claim.exclusivityAttestedAt ? (
            <Row
              label="Exclusivity attested"
              value={<TimestampDisplay value={claim.exclusivityAttestedAt} />}
            />
          ) : null}
        </dl>
      </CardContent>
    </Card>
  );
}
