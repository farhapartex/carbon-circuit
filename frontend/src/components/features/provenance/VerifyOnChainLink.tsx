import { ExternalLink, ShieldCheck } from "lucide-react";
import type { Checkpoint } from "@/lib/types";
import { explorerTransactionUrl } from "@/lib/wallet/chain";

type VerifyOnChainLinkProps = {
  checkpoints: Checkpoint[];
};

export function VerifyOnChainLink({ checkpoints }: VerifyOnChainLinkProps) {
  const anchored = checkpoints.filter(
    (checkpoint) => checkpoint.anchor.status === "confirmed",
  );

  if (anchored.length === 0) return null;

  const latest = anchored[anchored.length - 1];
  if (!latest?.anchor.transactionHash) return null;

  return (
    <section className="space-y-3 rounded-lg border border-neutral-200 bg-white p-6">
      <h2 className="inline-flex items-center gap-2 font-medium">
        <ShieldCheck className="size-4 text-primary-600" aria-hidden />
        Verify this yourself
      </h2>
      <p className="text-caption text-pretty text-neutral-600">
        {anchored.length} of {checkpoints.length} recorded steps are covered by
        a Merkle root written to a public ledger. You do not have to take this
        page&apos;s word for it: request the inclusion proof for any step and
        check it against the anchored root directly.
      </p>
      <a
        href={explorerTransactionUrl(latest.anchor.transactionHash)}
        target="_blank"
        rel="noopener noreferrer"
        className="inline-flex items-center gap-1.5 rounded-sm text-caption font-medium text-primary-700 underline underline-offset-4"
      >
        View the most recent anchor on the block explorer
        <ExternalLink className="size-3" aria-hidden />
      </a>
    </section>
  );
}
