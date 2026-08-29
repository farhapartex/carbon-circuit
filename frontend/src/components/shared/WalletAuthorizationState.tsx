"use client";

import { Wallet } from "lucide-react";
import type { ReactNode } from "react";
import { AddressDisplay } from "@/components/shared/AddressDisplay";
import { NetworkSwitcher } from "@/components/shared/NetworkSwitcher";
import { WalletConnectButton } from "@/components/shared/WalletConnectButton";
import { useWalletAuthorization } from "@/hooks/use-wallet-authorization";
import type { EthereumAddress } from "@/lib/types";

type WalletAuthorizationStateProps = {
  treasuryAddress: EthereumAddress | null;
  isApprovedOperator?: boolean | undefined;
  action: string;
  children: ReactNode;
};

export function WalletAuthorizationState({
  treasuryAddress,
  isApprovedOperator = false,
  action,
  children,
}: WalletAuthorizationStateProps) {
  const authorization = useWalletAuthorization(
    treasuryAddress,
    isApprovedOperator,
  );

  if (authorization.state === "authorized") {
    return <>{children}</>;
  }

  if (authorization.state === "wrong_network") {
    return <NetworkSwitcher />;
  }

  if (authorization.state === "disconnected") {
    return (
      <div className="flex flex-wrap items-center gap-3 rounded-lg bg-info-50 px-4 py-3">
        <Wallet className="size-4 text-info-600" aria-hidden />
        <p className="text-caption text-info-700">
          Connect a wallet to {action}.
        </p>
        <span className="ml-auto">
          <WalletConnectButton />
        </span>
      </div>
    );
  }

  return (
    <div className="space-y-2 rounded-lg bg-warning-50 px-4 py-3">
      <p className="text-caption text-warning-700">
        This wallet is not authorized to {action} for your organization. Credits
        are organization assets held at the Treasury Address, and a personal
        wallet is never their custodian.
      </p>
      <div className="flex flex-wrap items-center gap-x-6 gap-y-1 text-caption">
        <span className="flex items-center gap-2 text-warning-700">
          Connected
          <AddressDisplay
            address={authorization.connectedAddress}
            showCopy={false}
          />
        </span>
        {authorization.treasuryAddress ? (
          <span className="flex items-center gap-2 text-warning-700">
            Treasury
            <AddressDisplay address={authorization.treasuryAddress} />
          </span>
        ) : (
          <span className="text-warning-700">
            No Treasury Address is set for this organization yet.
          </span>
        )}
      </div>
    </div>
  );
}
