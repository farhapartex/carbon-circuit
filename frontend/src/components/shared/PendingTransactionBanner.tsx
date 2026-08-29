import { ExternalLink, Loader2 } from "lucide-react";
import type { TransactionHash } from "@/lib/types";
import { explorerTransactionUrl } from "@/lib/wallet/chain";
import { cn } from "@/lib/utils";

export type PendingOperation = {
  id: string;
  description: string;
  transactionHash: TransactionHash | null;
  submittedAt: string;
};

type PendingTransactionBannerProps = {
  operations: PendingOperation[];
  className?: string | undefined;
};

export function PendingTransactionBanner({
  operations,
  className,
}: PendingTransactionBannerProps) {
  if (operations.length === 0) return null;

  return (
    <section
      aria-live="polite"
      className={cn(
        "space-y-2 rounded-lg border border-info-600/30 bg-info-50 px-4 py-3",
        className,
      )}
    >
      <div className="flex items-center gap-2">
        <Loader2 className="size-4 animate-spin text-info-600" aria-hidden />
        <p className="text-caption font-medium text-info-700">
          {operations.length} operation
          {operations.length === 1 ? "" : "s"} still confirming on-chain
        </p>
      </div>
      <ul className="space-y-1">
        {operations.map((operation) => (
          <li
            key={operation.id}
            className="flex flex-wrap items-center gap-2 text-caption text-info-700"
          >
            {operation.description}
            {operation.transactionHash ? (
              <a
                href={explorerTransactionUrl(operation.transactionHash)}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1 rounded-sm underline underline-offset-4"
              >
                View
                <ExternalLink className="size-3" aria-hidden />
              </a>
            ) : null}
          </li>
        ))}
      </ul>
      <p className="text-caption text-info-700/80">
        These continue on the platform even if you close this tab. You will be
        notified when they settle.
      </p>
    </section>
  );
}
