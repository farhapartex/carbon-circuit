import type { Metadata } from "next";
import Link from "next/link";
import { BatchTable } from "@/components/features/provenance/BatchTable";
import { PageHeader } from "@/components/shared/PageHeader";
import { Button } from "@/components/ui/button";
import { fetchBatches } from "@/lib/api/batches";
import { auth0 } from "@/lib/auth0";

export const metadata: Metadata = { title: "Batches" };

export default async function BatchesPage() {
  const { token } = await auth0.getAccessToken();
  const page = await fetchBatches(token);

  return (
    <>
      <PageHeader
        title="Batches"
        description="Every produced quantity you register, and the journey recorded against it."
        actions={
          <Button asChild>
            <Link href="/batches/new">Create a batch</Link>
          </Button>
        }
      />

      <BatchTable
        batches={page.items}
        cursor={page.cursor}
        hasMore={page.hasMore}
      />
    </>
  );
}
