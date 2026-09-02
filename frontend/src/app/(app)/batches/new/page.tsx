import type { Metadata } from "next";
import { BatchWizard } from "@/components/features/provenance/BatchWizard";
import { PageHeader } from "@/components/shared/PageHeader";
import { fetchFacilities } from "@/lib/api/facilities";
import { fetchCurrentOrganization } from "@/lib/api/organization";
import { auth0 } from "@/lib/auth0";
import type { ProductCategory } from "@/lib/types";

export const metadata: Metadata = { title: "Create a batch" };

const PRODUCT_CATEGORIES: ProductCategory[] = [
  "electronics",
  "agriculture",
  "pharma",
  "textiles",
];

const declaredCategories = (declared: string[]): ProductCategory[] => {
  const recognised = declared.filter((category): category is ProductCategory =>
    PRODUCT_CATEGORIES.includes(category as ProductCategory),
  );

  return recognised.length > 0 ? recognised : ["electronics"];
};

export default async function NewBatchPage() {
  const { token } = await auth0.getAccessToken();

  const [organization, facilities] = await Promise.all([
    fetchCurrentOrganization(token),
    fetchFacilities(token),
  ]);

  return (
    <>
      <PageHeader
        backTo={{ href: "/batches", label: "Batches" }}
        title="Create a batch"
        description="Register a produced quantity of a component. Its journey is recorded against this record from here on."
      />

      <BatchWizard
        facilities={facilities}
        availableCategories={declaredCategories(organization.productCategories)}
      />
    </>
  );
}
