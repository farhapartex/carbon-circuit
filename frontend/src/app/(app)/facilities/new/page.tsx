import type { Metadata } from "next";
import { FacilityForm } from "@/components/features/facilities/FacilityForm";
import { PageHeader } from "@/components/shared/PageHeader";

export const metadata: Metadata = { title: "Add a facility" };

export default function NewFacilityPage() {
  return (
    <>
      <PageHeader
        backTo={{ href: "/facilities", label: "Facilities" }}
        title="Add a facility"
        description="Register a physical site under your organization. Its declared scale and registry match together determine how many credits it can ever earn."
      />

      <FacilityForm />
    </>
  );
}
