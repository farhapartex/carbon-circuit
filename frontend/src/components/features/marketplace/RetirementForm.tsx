"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { CreditAmountDisplay } from "@/components/shared/CreditAmountDisplay";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
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
  compareCreditAmounts,
  creditAmount,
  isZeroCreditAmount,
} from "@/lib/decimal";
import { activityTypeLabels } from "@/lib/labels";
import {
  RETIREMENT_PURPOSE_MINIMUM_LENGTH,
  type CreditClassBalance,
} from "@/lib/types";

const isPlainDecimal = (value: string) => /^\d+(\.\d{1,6})?$/.test(value);

const CONFIRM_PHRASE = "RETIRE";

export function RetirementForm({
  balances,
}: {
  balances: CreditClassBalance[];
}) {
  const router = useRouter();
  const retirable = balances.filter(
    (balance) => !isZeroCreditAmount(balance.available),
  );

  const [open, setOpen] = useState(false);
  const [tokenId, setTokenId] = useState(
    retirable[0]?.creditClass.tokenId ?? "",
  );
  const [quantity, setQuantity] = useState("");
  const [purpose, setPurpose] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const [notice, setNotice] = useState<string | null>(null);
  const [pending, startTransition] = useTransition();

  const selected = retirable.find(
    (balance) => balance.creditClass.tokenId === tokenId,
  );
  const parsed = isPlainDecimal(quantity) ? creditAmount(quantity) : null;

  const problem = (() => {
    if (!selected) return "Select a credit class with an available balance.";
    if (quantity !== "") {
      if (parsed === null) return "Quantity must be a plain decimal.";
      if (isZeroCreditAmount(parsed)) {
        return "Enter a quantity greater than zero.";
      }
      if (compareCreditAmounts(parsed, selected.available) > 0) {
        return "You cannot retire more than you hold in that class.";
      }
    }
    if (
      purpose !== "" &&
      purpose.trim().length < RETIREMENT_PURPOSE_MINIMUM_LENGTH
    ) {
      return `State the purpose in at least ${RETIREMENT_PURPOSE_MINIMUM_LENGTH} characters — it is published permanently.`;
    }
    return null;
  })();

  const ready =
    selected !== undefined &&
    parsed !== null &&
    purpose.trim().length >= RETIREMENT_PURPOSE_MINIMUM_LENGTH &&
    confirmation === CONFIRM_PHRASE &&
    problem === null;

  const submit = () => {
    startTransition(() => {
      setOpen(false);
      setNotice(
        "Retirements cannot be recorded yet — the marketplace service and chain writer are not built.",
      );
      router.refresh();
    });
  };

  return (
    <div className="space-y-4">
      {notice ? (
        <div
          role="alert"
          className="rounded-md border border-warning-600 bg-warning-50 px-4 py-3"
        >
          <p className="text-helper text-warning-700">{notice}</p>
        </div>
      ) : null}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogTrigger asChild>
          <Button disabled={retirable.length === 0}>Retire credits</Button>
        </DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Retire credits permanently</DialogTitle>
            <DialogDescription>
              Retiring marks credits as used to claim an offset. They can never
              be resold, re-transferred, or retired again by anyone, including
              you. This cannot be undone.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="retire-class">Credit class</Label>
              <Select value={tokenId} onValueChange={setTokenId}>
                <SelectTrigger id="retire-class" className="w-full">
                  <SelectValue placeholder="Select a credit class" />
                </SelectTrigger>
                <SelectContent>
                  {retirable.map((balance) => (
                    <SelectItem
                      key={balance.creditClass.tokenId}
                      value={balance.creditClass.tokenId}
                    >
                      {activityTypeLabels[balance.creditClass.activityType]},
                      vintage {balance.creditClass.vintageYear}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {selected ? (
                <p className="text-caption text-neutral-600">
                  Available: <CreditAmountDisplay amount={selected.available} />
                </p>
              ) : null}
            </div>

            <div className="space-y-2">
              <Label htmlFor="retire-quantity">Quantity, tCO2e</Label>
              <Input
                id="retire-quantity"
                inputMode="decimal"
                placeholder="250.000000"
                value={quantity}
                onChange={(event) => setQuantity(event.target.value)}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="retire-purpose">Purpose</Label>
              <Input
                id="retire-purpose"
                placeholder="2026 annual ESG report offset"
                value={purpose}
                onChange={(event) => setPurpose(event.target.value)}
              />
              <p className="text-caption text-neutral-600">
                Published on the public retirement log alongside your
                organization name, so write it for an outside reader.
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="retire-confirm">
                Type {CONFIRM_PHRASE} to confirm
              </Label>
              <Input
                id="retire-confirm"
                value={confirmation}
                onChange={(event) => setConfirmation(event.target.value)}
                autoComplete="off"
              />
            </div>

            {problem ? (
              <p
                role="alert"
                className="text-helper text-pretty text-danger-700"
              >
                {problem}
              </p>
            ) : null}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              disabled={!ready || pending}
              onClick={submit}
            >
              Retire permanently
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
