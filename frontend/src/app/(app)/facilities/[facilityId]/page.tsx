import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { FacilityTabs } from "@/components/features/facilities/FacilityTabs";
import { PageHeader } from "@/components/shared/PageHeader";
import {
  FacilityVerificationBadge,
  TrustTierBadge,
} from "@/components/shared/StatusBadges";
import { fetchFacility } from "@/lib/api/facilities";
import { GatewayError } from "@/lib/api/gateway";
import { auth0 } from "@/lib/auth0";
import { countryName } from "@/lib/countries";
import { facilityTypeLabels, gridRegionLabels } from "@/lib/labels";
import type { FacilityRecord } from "@/lib/api/facilities";

export const metadata: Metadata = { title: "Facility" };

const loadFacility = async (
  token: string,
  facilityId: string,
): Promise<FacilityRecord | null> => {
  try {
    return await fetchFacility(token, facilityId);
  } catch (error) {
    if (error instanceof GatewayError && error.status === 404) {
      return null;
    }
    throw error;
  }
};

export default async function FacilityDetailPage({
  params,
}: PageProps<"/facilities/[facilityId]">) {
  const { facilityId } = await params;
  const { token } = await auth0.getAccessToken();
  const facility = await loadFacility(token, facilityId);

  if (!facility) notFound();

  return (
    <>
      <PageHeader
        backTo={{ href: "/facilities", label: "Facilities" }}
        title={facility.name}
        description={`${facilityTypeLabels[facility.type]} in ${countryName(facility.countryCode)}, on the ${gridRegionLabels[facility.gridRegion]} grid.`}
        meta={
          <>
            <FacilityVerificationBadge status={facility.verificationStatus} />
            <TrustTierBadge tier={facility.trustTier} />
          </>
        }
      />

      <FacilityTabs facility={facility} />
    </>
  );
}
