"use client";

import { PenLine, ShieldCheck, Wallet } from "lucide-react";
import Link from "next/link";
import { useAccount } from "wagmi";
import { AddressDisplay } from "@/components/shared/AddressDisplay";
import { WalletConnectButton } from "@/components/shared/WalletConnectButton";
import { WalletProvider } from "@/components/shared/WalletProvider";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { EthereumAddress } from "@/lib/types";

const assurances = [
  {
    icon: ShieldCheck,
    title: "Credits belong to the organization",
    detail:
      "Not to whoever connected a wallet first. An employee leaving should never take the company's assets with them.",
  },
  {
    icon: PenLine,
    title: "You prove ownership by signing",
    detail:
      "We ask for a signature from this address over a single-use challenge. No transaction is submitted and nothing is spent.",
  },
  {
    icon: Wallet,
    title: "Changing it later takes 72 hours",
    detail:
      "With every Owner notified and able to cancel, so a compromised session cannot silently redirect your holdings.",
  },
];

function Designation() {
  const { address, isConnected } = useAccount();

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Treasury Address</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {isConnected && address ? (
            <>
              <div className="flex flex-wrap items-center gap-3 rounded-lg border border-neutral-200 px-4 py-3">
                <ShieldCheck
                  className="size-4 shrink-0 text-primary-600"
                  aria-hidden
                />
                <AddressDisplay address={address as EthereumAddress} />
              </div>
              <Button size="lg">Sign to designate this address</Button>
            </>
          ) : (
            <>
              <p className="text-caption text-pretty text-neutral-600">
                Connect the wallet your organization will use to hold credits.
              </p>
              <WalletConnectButton />
            </>
          )}
        </CardContent>
      </Card>

      <ul className="space-y-3">
        {assurances.map((assurance) => (
          <li key={assurance.title} className="flex gap-3">
            <assurance.icon
              className="mt-0.5 size-4 shrink-0 text-primary-600"
              aria-hidden
            />
            <span>
              <span className="block font-medium">{assurance.title}</span>
              <span className="block text-caption text-pretty text-neutral-600">
                {assurance.detail}
              </span>
            </span>
          </li>
        ))}
      </ul>

      <div className="flex flex-wrap items-center gap-4 border-t border-neutral-200 pt-6">
        <Button asChild variant="ghost">
          <Link href="/dashboard">Skip for now</Link>
        </Button>
        <p className="text-caption text-neutral-600">
          You can set this later in Settings, but no credits can be issued to
          your organization until you do.
        </p>
      </div>
    </div>
  );
}

export function TreasuryDesignation() {
  return (
    <WalletProvider>
      <Designation />
    </WalletProvider>
  );
}
