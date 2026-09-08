import type { Metadata } from "next";
import Link from "next/link";
import { ClaimsTable } from "@/components/features/claims/ClaimsTable";
import { PageHeader } from "@/components/shared/PageHeader";
import { Button } from "@/components/ui/button";
import { fetchClaims } from "@/lib/api/claims";
import { toSustainabilityClaim } from "@/lib/api/claimView";
import { auth0 } from "@/lib/auth0";

export const metadata: Metadata = { title: "Claims" };

export default async function ClaimsPage() {
  const { token } = await auth0.getAccessToken();
  const page = await fetchClaims(token);
  const claims = page.claims.map(toSustainabilityClaim);

  return (
    <>
      <PageHeader
        title="Sustainability claims"
        description="Each claim turns a facility's verified activity into carbon credits, capped by a computed ceiling."
        actions={
          <Button asChild>
            <Link href="/claims/new">Submit a claim</Link>
          </Button>
        }
      />

      <ClaimsTable claims={claims} />
    </>
  );
}
