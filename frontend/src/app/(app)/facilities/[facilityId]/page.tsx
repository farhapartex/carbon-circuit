import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { FacilityTabs } from "@/components/features/facilities/FacilityTabs";
import { PageHeader } from "@/components/shared/PageHeader";
import {
  FacilityVerificationBadge,
  TrustTierBadge,
} from "@/components/shared/StatusBadges";
import { countryName } from "@/lib/countries";
import {
  getCreditPortfolio,
  getFacility,
  listBatches,
  listClaims,
} from "@/lib/fixtures";
import { facilityTypeLabels } from "@/lib/labels";

export const metadata: Metadata = { title: "Facility" };

export default async function FacilityDetailPage({
  params,
}: PageProps<"/facilities/[facilityId]">) {
  const { facilityId } = await params;
  const facility = await getFacility(facilityId);

  if (!facility) notFound();

  const [batches, claims, portfolio] = await Promise.all([
    listBatches(),
    listClaims(),
    getCreditPortfolio(),
  ]);

  return (
    <>
      <PageHeader
        backTo={{ href: "/facilities", label: "Facilities" }}
        title={facility.name}
        description={`${facilityTypeLabels[facility.type]} in ${countryName(facility.countryCode)}, on the ${facility.gridRegion} grid.`}
        meta={
          <>
            <FacilityVerificationBadge status={facility.verificationStatus} />
            <TrustTierBadge tier={facility.trustTier} />
          </>
        }
      />

      <FacilityTabs
        facility={facility}
        batches={batches.items.filter(
          (batch) => batch.originatingFacilityId === facility.id,
        )}
        claims={claims.items.filter(
          (claim) => claim.facilityId === facility.id,
        )}
        balances={portfolio.balances.filter(
          (balance) => balance.creditClass.facilityId === facility.id,
        )}
      />
    </>
  );
}
