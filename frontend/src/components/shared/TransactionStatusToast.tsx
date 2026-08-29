"use client";

import { ExternalLink } from "lucide-react";
import { TransactionStatusPill } from "@/components/shared/StatusBadges";
import type { TransactionStatus } from "@/lib/status";
import type { TransactionHash } from "@/lib/types";
import { explorerTransactionUrl } from "@/lib/wallet/chain";
import { cn } from "@/lib/utils";

type TransactionStatusToastProps = {
  status: TransactionStatus;
  description: string;
  transactionHash?: TransactionHash | undefined;
  className?: string | undefined;
};

export function TransactionStatusToast({
  status,
  description,
  transactionHash,
  className,
}: TransactionStatusToastProps) {
  return (
    <div
      data-slot="transaction-status"
      role="status"
      aria-live="polite"
      className={cn(
        "flex flex-wrap items-center gap-3 rounded-lg border border-neutral-200 bg-white px-4 py-3",
        className,
      )}
    >
      <TransactionStatusPill status={status} />
      <p className="text-caption text-muted-foreground">{description}</p>
      {transactionHash ? (
        <a
          href={explorerTransactionUrl(transactionHash)}
          target="_blank"
          rel="noopener noreferrer"
          className="ml-auto inline-flex items-center gap-1 rounded-sm text-caption font-medium text-primary-700 underline underline-offset-4"
        >
          View on explorer
          <ExternalLink className="size-3" aria-hidden />
        </a>
      ) : null}
    </div>
  );
}
