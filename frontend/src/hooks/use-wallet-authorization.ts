"use client";

import { useAccount, useChainId } from "wagmi";
import type { EthereumAddress } from "@/lib/types";
import { expectedChain } from "@/lib/wallet/chain";

export type WalletAuthorization =
  | { state: "disconnected" }
  | { state: "wrong_network"; connectedChainId: number }
  | {
      state: "unauthorized";
      connectedAddress: EthereumAddress;
      treasuryAddress: EthereumAddress | null;
    }
  | { state: "authorized"; connectedAddress: EthereumAddress };

const sameAddress = (left: string | undefined, right: string | null) =>
  Boolean(left && right && left.toLowerCase() === right.toLowerCase());

export function useWalletAuthorization(
  treasuryAddress: EthereumAddress | null,
  isApprovedOperator = false,
): WalletAuthorization {
  const { address, isConnected } = useAccount();
  const connectedChainId = useChainId();

  if (!isConnected || !address) {
    return { state: "disconnected" };
  }

  if (connectedChainId !== expectedChain.id) {
    return { state: "wrong_network", connectedChainId };
  }

  if (sameAddress(address, treasuryAddress) || isApprovedOperator) {
    return { state: "authorized", connectedAddress: address };
  }

  return { state: "unauthorized", connectedAddress: address, treasuryAddress };
}
