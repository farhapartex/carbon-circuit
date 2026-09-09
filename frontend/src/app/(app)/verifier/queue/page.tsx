import type { Metadata } from "next";
import { ReviewQueueTable } from "@/components/features/verifier/ReviewQueueTable";
import { PageHeader } from "@/components/shared/PageHeader";
import { fetchReviewQueue } from "@/lib/api/verifier";
import { auth0 } from "@/lib/auth0";

export const metadata: Metadata = { title: "Review queue" };

export default async function VerifierQueuePage() {
  const { token } = await auth0.getAccessToken();
  const claims = await fetchReviewQueue(token);

  return (
    <>
      <PageHeader
        title="Review queue"
        description="Claims awaiting a verification decision, most urgent first, then oldest."
      />

      <ReviewQueueTable claims={claims} />
    </>
  );
}
