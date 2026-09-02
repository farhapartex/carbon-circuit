import type { Metadata } from "next";
import Link from "next/link";
import { FacilityTable } from "@/components/features/facilities/FacilityTable";
import { PageHeader } from "@/components/shared/PageHeader";
import { Button } from "@/components/ui/button";
import { listFacilities } from "@/lib/fixtures";

export const metadata: Metadata = { title: "Facilities" };

export default async function FacilitiesPage() {
  const facilities = await listFacilities();

  return (
    <>
      <PageHeader
        title="Facilities"
        description="Physical sites that produce batches and earn carbon credits."
        actions={
          <Button asChild>
            <Link href="/facilities/new">Add a facility</Link>
          </Button>
        }
      />

      <FacilityTable facilities={facilities.items} />
    </>
  );
}
