"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { EscrowExplainer } from "@/components/features/marketplace/EscrowExplainer";
import { PricePreview } from "@/components/features/marketplace/PricePreview";
import { CreditAmountDisplay } from "@/components/shared/CreditAmountDisplay";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  applyBasisPoints,
  compareCreditAmounts,
  compareUsdcAmounts,
  costOf,
  creditAmount,
  isZeroCreditAmount,
  subtractUsdcAmounts,
  usdcAmount,
} from "@/lib/decimal";
import { activityTypeLabels } from "@/lib/labels";
import {
  LISTING_MAXIMUM_DURATION_DAYS,
  MAXIMUM_PRICE_PER_TONNE_USDC,
  MINIMUM_LISTING_QUANTITY,
  MINIMUM_PRICE_PER_TONNE_USDC,
  MINIMUM_PURCHASE_QUANTITY_FLOOR,
  type CreditClassBalance,
} from "@/lib/types";

const isPlainDecimal = (value: string) => /^\d+(\.\d{1,6})?$/.test(value);

const dayOffset = (days: number) =>
  new Date(Date.now() + days * 86_400_000).toISOString().slice(0, 10);

type ListingFormProps = {
  balances: CreditClassBalance[];
  feeBasisPoints: number | null;
};

export function ListingForm({ balances, feeBasisPoints }: ListingFormProps) {
  const router = useRouter();
  const sellable = balances.filter(
    (balance) => !isZeroCreditAmount(balance.available),
  );

  const [tokenId, setTokenId] = useState(
    sellable[0]?.creditClass.tokenId ?? "",
  );
  const [quantity, setQuantity] = useState("");
  const [price, setPrice] = useState("");
  const [minimumPurchase, setMinimumPurchase] = useState("");
  const [expiresOn, setExpiresOn] = useState(
    dayOffset(LISTING_MAXIMUM_DURATION_DAYS),
  );
  const [notice, setNotice] = useState<string | null>(null);
  const [pending, startTransition] = useTransition();

  const selected = sellable.find(
    (balance) => balance.creditClass.tokenId === tokenId,
  );

  const parsedQuantity = isPlainDecimal(quantity)
    ? creditAmount(quantity)
    : null;
  const parsedPrice = isPlainDecimal(price) ? usdcAmount(price) : null;
  const parsedMinimum = isPlainDecimal(minimumPurchase)
    ? creditAmount(minimumPurchase)
    : null;

  const gross =
    parsedQuantity === null || parsedPrice === null
      ? null
      : costOf(parsedQuantity, parsedPrice);
  const fee =
    gross === null || feeBasisPoints === null
      ? null
      : applyBasisPoints(gross, feeBasisPoints);
  const net =
    gross === null || fee === null ? null : subtractUsdcAmounts(gross, fee);

  const problem = (() => {
    if (!selected) return "Select a credit class with an available balance.";

    if (quantity !== "") {
      if (parsedQuantity === null) {
        return "Quantity must be a plain number with up to six decimal places.";
      }
      if (
        compareCreditAmounts(
          parsedQuantity,
          creditAmount(MINIMUM_LISTING_QUANTITY),
        ) < 0
      ) {
        return `The minimum listing quantity is ${MINIMUM_LISTING_QUANTITY} tCO2e.`;
      }
      if (compareCreditAmounts(parsedQuantity, selected.available) > 0) {
        return "You cannot list more than you have available in that class.";
      }
    }

    if (price !== "") {
      if (parsedPrice === null) return "Price must be a plain decimal.";
      if (
        compareUsdcAmounts(
          parsedPrice,
          usdcAmount(MINIMUM_PRICE_PER_TONNE_USDC),
        ) < 0 ||
        compareUsdcAmounts(
          parsedPrice,
          usdcAmount(MAXIMUM_PRICE_PER_TONNE_USDC),
        ) > 0
      ) {
        return `Price must be between ${MINIMUM_PRICE_PER_TONNE_USDC} and ${MAXIMUM_PRICE_PER_TONNE_USDC} USDC per tCO2e.`;
      }
    }

    if (minimumPurchase !== "") {
      if (parsedMinimum === null) {
        return "Minimum purchase must be a plain decimal.";
      }
      if (
        compareCreditAmounts(
          parsedMinimum,
          creditAmount(MINIMUM_PURCHASE_QUANTITY_FLOOR),
        ) < 0
      ) {
        return `The minimum purchase quantity cannot be below ${MINIMUM_PURCHASE_QUANTITY_FLOOR} tCO2e.`;
      }
      if (
        parsedQuantity !== null &&
        compareCreditAmounts(parsedMinimum, parsedQuantity) > 0
      ) {
        return "The minimum purchase cannot exceed the quantity you are listing.";
      }
    }

    if (expiresOn > dayOffset(LISTING_MAXIMUM_DURATION_DAYS)) {
      return `A listing can run for at most ${LISTING_MAXIMUM_DURATION_DAYS} days.`;
    }

    return null;
  })();

  const complete =
    selected !== undefined &&
    parsedQuantity !== null &&
    parsedPrice !== null &&
    parsedMinimum !== null &&
    problem === null;

  const submit = () => {
    setNotice(null);
    startTransition(() => {
      setNotice(
        "Listings cannot be created yet — the marketplace service and chain writer are not built.",
      );
      router.refresh();
    });
  };

  if (sellable.length === 0) {
    return (
      <p className="rounded-lg border border-dashed border-neutral-200 bg-white px-6 py-10 text-center text-caption text-pretty text-neutral-600">
        You have no credits available to list. Credits arrive once a verifier
        approves a sustainability claim for one of your facilities.
      </p>
    );
  }

  return (
    <div className="max-w-xl space-y-6">
      {notice ? (
        <div
          role="alert"
          className="rounded-md border border-warning-600 bg-warning-50 px-4 py-3"
        >
          <p className="text-helper text-warning-700">{notice}</p>
        </div>
      ) : null}

      <div className="space-y-2">
        <Label htmlFor="credit-class">Credit class</Label>
        <Select value={tokenId} onValueChange={setTokenId}>
          <SelectTrigger id="credit-class" className="w-full">
            <SelectValue placeholder="Select a credit class" />
          </SelectTrigger>
          <SelectContent>
            {sellable.map((balance) => (
              <SelectItem
                key={balance.creditClass.tokenId}
                value={balance.creditClass.tokenId}
              >
                {activityTypeLabels[balance.creditClass.activityType]},{" "}
                {balance.creditClass.facilityName}, vintage{" "}
                {balance.creditClass.vintageYear}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {selected ? (
          <p className="text-caption text-neutral-600">
            Available in this class:{" "}
            <CreditAmountDisplay amount={selected.available} />
          </p>
        ) : null}
        <p className="text-caption text-neutral-600">
          A listing sells from exactly one credit class, because a credit
          permanently carries its facility, vintage, and activity.
        </p>
      </div>

      <div className="space-y-2">
        <Label htmlFor="listing-quantity">Quantity, tCO2e</Label>
        <Input
          id="listing-quantity"
          inputMode="decimal"
          placeholder="500.000000"
          value={quantity}
          onChange={(event) => setQuantity(event.target.value)}
        />
      </div>

      <div className="grid gap-6 sm:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="listing-price">Price per tCO2e, USDC</Label>
          <Input
            id="listing-price"
            inputMode="decimal"
            placeholder="12.50"
            value={price}
            onChange={(event) => setPrice(event.target.value)}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="listing-minimum">Minimum purchase, tCO2e</Label>
          <Input
            id="listing-minimum"
            inputMode="decimal"
            placeholder="10.000000"
            value={minimumPurchase}
            onChange={(event) => setMinimumPurchase(event.target.value)}
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="listing-expiry">Expires on</Label>
        <Input
          id="listing-expiry"
          type="date"
          min={dayOffset(1)}
          max={dayOffset(LISTING_MAXIMUM_DURATION_DAYS)}
          value={expiresOn}
          onChange={(event) => setExpiresOn(event.target.value)}
        />
        <p className="text-caption text-neutral-600">
          At expiry, anything unsold returns to your available balance
          automatically.
        </p>
      </div>

      {problem ? (
        <p role="alert" className="text-helper text-pretty text-danger-700">
          {problem}
        </p>
      ) : null}

      <PricePreview
        quantity={parsedQuantity}
        gross={gross}
        fee={fee}
        net={net}
        feeBasisPoints={feeBasisPoints}
      />

      <EscrowExplainer />

      <Button
        type="button"
        size="lg"
        disabled={!complete || pending}
        onClick={submit}
      >
        Create listing
      </Button>
    </div>
  );
}
