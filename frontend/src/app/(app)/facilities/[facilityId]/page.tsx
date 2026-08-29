import { PageHeader } from "@/components/shared/PageHeader";

export default async function FacilitiesFacilityidPage(
  props: PageProps<"/facilities/[facilityId]">,
) {
  const { facilityId } = await props.params;
  return <PageHeader title="Facility" description={facilityId} />;
}
