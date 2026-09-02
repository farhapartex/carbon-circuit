import type { Metadata } from "next";
import { BatchWizard } from "@/components/features/provenance/BatchWizard";
import { PageHeader } from "@/components/shared/PageHeader";
import { getCurrentOrganization, listFacilities } from "@/lib/fixtures";

export const metadata: Metadata = { title: "Create a batch" };

export default async function NewBatchPage() {
  const [organization, facilities] = await Promise.all([
    getCurrentOrganization(),
    listFacilities(),
  ]);

  return (
    <>
      <PageHeader
        backTo={{ href: "/batches", label: "Batches" }}
        title="Create a batch"
        description="Register a produced quantity of a component. Its journey is recorded against this record from here on."
      />

      <BatchWizard
        facilities={facilities.items}
        availableCategories={
          organization.productCategories.length > 0
            ? organization.productCategories
            : ["electronics"]
        }
      />
    </>
  );
}
