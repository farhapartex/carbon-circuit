import type { Metadata } from "next";
import { Clock, ShieldCheck } from "lucide-react";
import { AddressDisplay } from "@/components/shared/AddressDisplay";
import { CreditAmountDisplay } from "@/components/shared/CreditAmountDisplay";
import { StatusPill } from "@/components/shared/StatusPill";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  getCreditPortfolio,
  getCurrentOrganization,
  getPendingTreasuryChange,
} from "@/lib/fixtures";

export const metadata: Metadata = { title: "Treasury Address" };

export default async function SettingsWalletPage() {
  const organization = await getCurrentOrganization();
  const portfolio = await getCreditPortfolio();
  const pendingChange = await getPendingTreasuryChange();
  const hasEscrow = portfolio.totalEscrowed !== "0";

  return (
    <>
      {pendingChange.state === "pending" ? (
        <section
          role="alert"
          className="space-y-3 rounded-lg border border-warning-600/30 bg-warning-50 px-6 py-4"
        >
          <p className="flex items-center gap-2 font-medium text-warning-700">
            <Clock className="size-4" aria-hidden />
            Treasury Address change pending
          </p>
          <p className="text-caption text-pretty text-warning-700">
            {pendingChange.initiatedBy} requested a change to{" "}
            <AddressDisplay
              address={pendingChange.requestedAddress}
              showCopy={false}
            />{" "}
            on <TimestampDisplay value={pendingChange.initiatedAt} dateOnly />.
            It takes effect{" "}
            <TimestampDisplay value={pendingChange.effectiveAt} />. Any Owner
            can cancel it until then.
          </p>
          <Button variant="outline" size="sm">
            Cancel this change
          </Button>
        </section>
      ) : null}

      <Card>
        <CardHeader className="flex-row items-start justify-between gap-4">
          <CardTitle>Treasury Address</CardTitle>
          <StatusPill
            presentation={
              organization.treasuryAddress
                ? { label: "Designated", variant: "success" }
                : { label: "Not set", variant: "warning" }
            }
          />
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-caption text-pretty text-neutral-600">
            This single address receives, holds, and spends your
            organization&apos;s carbon credits. Credits belong to the
            organization rather than to whichever employee connected a wallet
            first — minting to a personal address would mean an employee leaving
            takes the company&apos;s assets with them.
          </p>

          {organization.treasuryAddress ? (
            <>
              <div className="flex flex-wrap items-center gap-3 rounded-lg border border-neutral-200 px-4 py-3">
                <ShieldCheck
                  className="size-4 shrink-0 text-primary-600"
                  aria-hidden
                />
                <AddressDisplay address={organization.treasuryAddress} />
              </div>

              <dl className="grid gap-4 sm:grid-cols-2">
                <div>
                  <dt className="text-caption text-neutral-600">
                    Credits held
                  </dt>
                  <dd className="font-medium">
                    <CreditAmountDisplay amount={portfolio.totalAvailable} />
                  </dd>
                </div>
                <div>
                  <dt className="text-caption text-neutral-600">
                    In escrow on listings
                  </dt>
                  <dd className="font-medium">
                    <CreditAmountDisplay amount={portfolio.totalEscrowed} />
                  </dd>
                </div>
              </dl>
            </>
          ) : (
            <Button>Connect a wallet</Button>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Change the Treasury Address</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-caption text-pretty text-neutral-600">
            Changing it requires an Owner, a fresh signature from the new
            address, multi-factor re-authentication, and a 72-hour delay during
            which every Owner and administrator is notified and any Owner can
            cancel. That delay exists so an attacker who compromises one session
            cannot silently redirect your entire credit holding.
          </p>

          {hasEscrow ? (
            <p className="rounded-lg bg-warning-50 px-4 py-3 text-caption text-warning-700">
              You have <CreditAmountDisplay amount={portfolio.totalEscrowed} />{" "}
              in escrow on active listings. Cancel or fill those listings before
              changing the address.
            </p>
          ) : null}

          <Button variant="outline" disabled={hasEscrow}>
            Start a Treasury Address change
          </Button>
        </CardContent>
      </Card>
    </>
  );
}
