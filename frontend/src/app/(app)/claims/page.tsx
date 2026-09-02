import type { Metadata } from "next";
import Link from "next/link";
import { ClaimsTable } from "@/components/features/claims/ClaimsTable";
import { PageHeader } from "@/components/shared/PageHeader";
import { Button } from "@/components/ui/button";
import { listClaims } from "@/lib/fixtures";

export const metadata: Metadata = { title: "Claims" };

export default async function ClaimsPage() {
  const claims = await listClaims();

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

      <ClaimsTable claims={claims.items} />
    </>
  );
}
