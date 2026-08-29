"use client";

import { ConnectButton } from "@rainbow-me/rainbowkit";
import { Wallet } from "lucide-react";
import { AddressDisplay } from "@/components/shared/AddressDisplay";
import { Button } from "@/components/ui/button";
import type { EthereumAddress } from "@/lib/types";

export function WalletConnectButton() {
  return (
    <ConnectButton.Custom>
      {({
        account,
        chain,
        openAccountModal,
        openChainModal,
        openConnectModal,
        mounted,
      }) => {
        if (!mounted) {
          return <Button variant="outline" disabled aria-hidden />;
        }

        if (!account || !chain) {
          return (
            <Button variant="outline" onClick={openConnectModal}>
              <Wallet className="size-4" aria-hidden />
              Connect wallet
            </Button>
          );
        }

        if (chain.unsupported) {
          return (
            <Button variant="destructive" onClick={openChainModal}>
              Wrong network
            </Button>
          );
        }

        return (
          <Button variant="outline" onClick={openAccountModal}>
            <AddressDisplay
              address={account.address as EthereumAddress}
              showCopy={false}
            />
          </Button>
        );
      }}
    </ConnectButton.Custom>
  );
}
