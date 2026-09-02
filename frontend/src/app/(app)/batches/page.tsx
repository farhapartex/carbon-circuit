import type { Metadata } from "next";
import Link from "next/link";
import { BatchTable } from "@/components/features/provenance/BatchTable";
import { PageHeader } from "@/components/shared/PageHeader";
import { Button } from "@/components/ui/button";
import { listBatches } from "@/lib/fixtures";

export const metadata: Metadata = { title: "Batches" };

export default async function BatchesPage() {
  const batches = await listBatches();

  return (
    <>
      <div className="flex flex-wrap items-start justify-between gap-4">
        <PageHeader
          title="Batches"
          description="Every produced quantity your organization has registered, and how far each one has been traced."
        />
        <Button asChild>
          <Link href="/batches/new">Register a batch</Link>
        </Button>
      </div>

      <BatchTable batches={batches.items} meta={batches.meta} />
    </>
  );
}
