import type { Metadata } from "next";
import Link from "next/link";
import { FacilityTable } from "@/components/features/facilities/FacilityTable";
import { PageHeader } from "@/components/shared/PageHeader";
import { Button } from "@/components/ui/button";
import { fetchFacilities } from "@/lib/api/facilities";
import { auth0 } from "@/lib/auth0";

export const metadata: Metadata = { title: "Facilities" };

export default async function FacilitiesPage() {
  const { token } = await auth0.getAccessToken();
  const facilities = await fetchFacilities(token);

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

      <FacilityTable facilities={facilities} />
    </>
  );
}
