import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { discountFactorRationale } from "@/lib/labels";
import type { Facility } from "@/lib/types";

const decimalFormat = new Intl.NumberFormat("en-US", {
  maximumFractionDigits: 0,
});

const format = (value: string | null) =>
  value === null ? null : decimalFormat.format(Number(value));

type FigureProps = {
  label: string;
  declared: string;
  attested: string | null;
};

function Figure({ label, declared, attested }: FigureProps) {
  const attestedDisplay = format(attested);

  return (
    <div className="space-y-1">
      <p className="text-caption text-neutral-600">{label}</p>
      <p className="font-medium tabular-nums">{format(declared)}</p>
      <p className="text-caption text-neutral-600">
        {attestedDisplay === null
          ? "No attested figure on record"
          : `Attested ${attestedDisplay}`}
      </p>
    </div>
  );
}

export function CapacityCard({ facility }: { facility: Facility }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Capacity and ceiling discount</CardTitle>
      </CardHeader>
      <CardContent className="space-y-5">
        <div className="grid gap-4 sm:grid-cols-2">
          <Figure
            label="Annual production capacity"
            declared={facility.declaredAnnualProductionCapacity}
            attested={facility.attestedAnnualProductionCapacity}
          />
          <Figure
            label="Annual energy consumption, kWh"
            declared={facility.declaredAnnualEnergyConsumptionKwh}
            attested={facility.attestedAnnualEnergyConsumptionKwh}
          />
        </div>

        <div className="rounded-md border border-neutral-200 bg-neutral-50 px-4 py-3">
          <p className="flex items-baseline justify-between gap-4">
            <span className="text-caption text-neutral-600">
              Ceiling discount factor
            </span>
            <span className="text-lg font-medium tabular-nums">
              {facility.ceilingDiscountFactor}
            </span>
          </p>
          <p className="mt-2 text-caption text-pretty text-neutral-600">
            {discountFactorRationale[facility.verificationStatus]}
          </p>
        </div>
      </CardContent>
    </Card>
  );
}
