import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { countryName } from "@/lib/countries";
import { facilityTypeLabels } from "@/lib/labels";
import type { FacilityRecord } from "@/lib/api/facilities";

export function FacilityProfileCard({
  facility,
}: {
  facility: FacilityRecord;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Site</CardTitle>
      </CardHeader>
      <CardContent>
        <dl className="grid gap-3 sm:grid-cols-2">
          <div className="sm:col-span-2">
            <dt className="text-caption text-neutral-600">Address</dt>
            <dd className="font-medium">{facility.address}</dd>
          </div>
          <div>
            <dt className="text-caption text-neutral-600">Country</dt>
            <dd className="font-medium">{countryName(facility.countryCode)}</dd>
          </div>
          <div>
            <dt className="text-caption text-neutral-600">Type</dt>
            <dd className="font-medium">{facilityTypeLabels[facility.type]}</dd>
          </div>
          <div>
            <dt className="text-caption text-neutral-600">Grid region</dt>
            <dd className="font-mono text-helper">{facility.gridRegion}</dd>
          </div>
          <div>
            <dt className="text-caption text-neutral-600">Registered</dt>
            <dd className="font-medium">
              <TimestampDisplay value={facility.createdAt} dateOnly />
            </dd>
          </div>
        </dl>
      </CardContent>
    </Card>
  );
}
