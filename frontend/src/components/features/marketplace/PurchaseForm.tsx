"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { CreditAmountDisplay } from "@/components/shared/CreditAmountDisplay";
import { UsdcAmountDisplay } from "@/components/shared/UsdcAmountDisplay";
import { WalletAuthorizationState } from "@/components/shared/WalletAuthorizationState";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  compareCreditAmounts,
  compareUsdcAmounts,
  costOf,
  creditAmount,
  isZeroCreditAmount,
  subtractCreditAmounts,
  usdcAmount,
} from "@/lib/decimal";
import {
  MINIMUM_TRANSACTION_NOTIONAL_USDC,
  type EthereumAddress,
  type MarketplaceListing,
} from "@/lib/types";

const isPlainQuantity = (value: string) => /^\d+(\.\d{1,6})?$/.test(value);

type PurchaseFormProps = {
  listing: MarketplaceListing;
  treasuryAddress: EthereumAddress | null;
};

export function PurchaseForm({ listing, treasuryAddress }: PurchaseFormProps) {
  const router = useRouter();
  const [quantity, setQuantity] = useState("");
  const [failure, setFailure] = useState<string | null>(null);
  const [pending, startTransition] = useTransition();

  const parsed = isPlainQuantity(quantity) ? creditAmount(quantity) : null;
  const notionalFloor = usdcAmount(MINIMUM_TRANSACTION_NOTIONAL_USDC);

  const cost = parsed === null ? null : costOf(parsed, listing.pricePerTonne);

  const problem = (() => {
    if (quantity === "") return null;
    if (parsed === null) {
      return "Enter a plain number with up to six decimal places.";
    }
    if (isZeroCreditAmount(parsed)) {
      return "Enter a quantity greater than zero.";
    }
    if (compareCreditAmounts(parsed, listing.minimumPurchaseQuantity) < 0) {
      return "That is below this listing's minimum purchase quantity.";
    }
    if (compareCreditAmounts(parsed, listing.quantityRemaining) > 0) {
      return "That is more than the listing has remaining.";
    }

    const remainder = subtractCreditAmounts(listing.quantityRemaining, parsed);
    if (
      !isZeroCreditAmount(remainder) &&
      compareCreditAmounts(remainder, listing.minimumPurchaseQuantity) < 0
    ) {
      return "This would leave a remainder smaller than the listing's own minimum, which nobody could then buy. Either buy the listing out entirely, or leave at least the minimum behind.";
    }

    if (cost !== null && compareUsdcAmounts(cost, notionalFloor) < 0) {
      return `A purchase must come to at least ${MINIMUM_TRANSACTION_NOTIONAL_USDC} USDC.`;
    }

    return null;
  })();

  const purchasable = parsed !== null && problem === null;

  const submit = () => {
    setFailure(null);
    startTransition(() => {
      setFailure(
        "Purchases cannot be settled yet — the marketplace service and chain writer are not built.",
      );
      router.refresh();
    });
  };

  return (
    <div className="space-y-4 rounded-lg border border-neutral-200 bg-white p-5">
      <h2 className="font-medium">Buy from this listing</h2>

      {failure ? (
        <div
          role="alert"
          className="rounded-md border border-warning-600 bg-warning-50 px-4 py-3"
        >
          <p className="text-helper text-warning-700">{failure}</p>
        </div>
      ) : null}

      <div className="space-y-2">
        <Label htmlFor="purchase-quantity">Quantity, tCO2e</Label>
        <Input
          id="purchase-quantity"
          inputMode="decimal"
          placeholder="100.000000"
          value={quantity}
          onChange={(event) => setQuantity(event.target.value)}
          aria-invalid={problem !== null}
        />
        <p className="text-caption text-neutral-600">
          At least{" "}
          <CreditAmountDisplay amount={listing.minimumPurchaseQuantity} />, at
          most <CreditAmountDisplay amount={listing.quantityRemaining} />.
        </p>
        {problem ? (
          <p role="alert" className="text-helper text-pretty text-danger-700">
            {problem}
          </p>
        ) : null}
      </div>

      <dl className="space-y-2 border-t border-neutral-200 pt-4">
        <div className="flex flex-wrap items-baseline justify-between gap-4">
          <dt className="text-caption text-neutral-600">Price per tCO2e</dt>
          <dd className="font-medium">
            <UsdcAmountDisplay amount={listing.pricePerTonne} />
          </dd>
        </div>
        <div className="flex flex-wrap items-baseline justify-between gap-4">
          <dt className="text-caption text-neutral-600">Platform fee</dt>
          <dd className="font-medium">None — buyers pay no fee</dd>
        </div>
        <div className="flex flex-wrap items-baseline justify-between gap-4 border-t border-neutral-200 pt-2">
          <dt className="font-medium">Total</dt>
          <dd className="text-lg font-medium">
            {cost === null ? (
              <span className="text-neutral-600">—</span>
            ) : (
              <UsdcAmountDisplay amount={cost} exact />
            )}
          </dd>
        </div>
      </dl>

      <WalletAuthorizationState
        treasuryAddress={treasuryAddress}
        action="purchase credits"
      >
        <Button
          type="button"
          size="lg"
          className="w-full"
          disabled={!purchasable || pending}
          onClick={submit}
        >
          Buy credits
        </Button>
      </WalletAuthorizationState>
    </div>
  );
}
