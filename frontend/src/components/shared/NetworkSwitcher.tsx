"use client";

import { AlertTriangle } from "lucide-react";
import { useSwitchChain } from "wagmi";
import { Button } from "@/components/ui/button";
import { expectedChain } from "@/lib/wallet/chain";

export function NetworkSwitcher() {
  const { switchChain, isPending } = useSwitchChain();

  return (
    <div
      role="alert"
      className="flex flex-wrap items-center gap-3 rounded-lg bg-warning-50 px-4 py-3"
    >
      <AlertTriangle className="size-4 text-warning-600" aria-hidden />
      <p className="text-caption text-warning-700">
        Your wallet is connected to a different network. Switch to{" "}
        {expectedChain.name} to continue.
      </p>
      <Button
        size="sm"
        variant="outline"
        className="ml-auto"
        disabled={isPending}
        onClick={() => switchChain({ chainId: expectedChain.id })}
      >
        {isPending ? "Switching…" : `Switch to ${expectedChain.name}`}
      </Button>
    </div>
  );
}
