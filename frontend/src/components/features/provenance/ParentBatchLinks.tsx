import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { countryName } from "@/lib/countries";
import type { Id, ParentBatchReference } from "@/lib/types";

type ParentBatchLinksProps = {
  batchId: Id;
  parents: ParentBatchReference[];
};

export function ParentBatchLinks({ batchId, parents }: ParentBatchLinksProps) {
  if (parents.length === 0) return null;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Component batches</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {parents.map((parent) => (
          <div
            key={parent.id}
            className="flex flex-wrap items-center justify-between gap-2 rounded-md border border-neutral-200 px-4 py-3"
          >
            <span>
              <span className="block font-medium">{parent.componentType}</span>
              <span className="block text-caption text-neutral-600">
                {parent.originatingFacilityName} ·{" "}
                {countryName(parent.originatingFacilityCountry)}
              </span>
            </span>
            {parent.resolved ? (
              <Button asChild size="sm" variant="outline">
                <Link href={`/batches/${batchId}/components/${parent.id}`}>
                  View
                </Link>
              </Button>
            ) : (
              <span className="text-caption text-neutral-600">
                Held by another organization
              </span>
            )}
          </div>
        ))}
      </CardContent>
    </Card>
  );
}
