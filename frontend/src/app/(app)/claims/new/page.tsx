import type { Metadata } from "next";
import { ClaimWizard } from "@/components/features/claims/ClaimWizard";
import { PageHeader } from "@/components/shared/PageHeader";
import { fetchFacilities } from "@/lib/api/facilities";
import { auth0 } from "@/lib/auth0";

export const metadata: Metadata = { title: "Submit a claim" };

export default async function NewClaimPage() {
  const session = await auth0.getSession();
  const { token } = await auth0.getAccessToken();
  const facilities = await fetchFacilities(token);

  return (
    <>
      <PageHeader
        backTo={{ href: "/claims", label: "Claims" }}
        title="Submit a sustainability claim"
        description="A claim describes a sustainability practice over a defined period. Its credit ceiling is computed from your facility's attested scale, never from the figure you request."
      />

      <ClaimWizard
        facilities={facilities}
        userName={session?.user.name ?? session?.user.email ?? "this user"}
      />
    </>
  );
}
