"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { recordDecision } from "@/lib/actions/verifier";
import type { DecisionOutcome } from "@/lib/api/verifier";

const MINIMUM_REASON = 40;

type DecisionActionBarProps = {
  claimId: string;
  requestedAmount: string;
  computedCeiling: string;
  requiresDualApproval: boolean;
  approvalsSoFar: number;
};

const lowerOf = (first: string, second: string) =>
  Number(first) < Number(second) ? first : second;

export function DecisionActionBar({
  claimId,
  requestedAmount,
  computedCeiling,
  requiresDualApproval,
  approvalsSoFar,
}: DecisionActionBarProps) {
  const router = useRouter();
  const [pending, startTransition] = useTransition();
  const [idempotencyKey] = useState(() => crypto.randomUUID());

  const issuable = lowerOf(requestedAmount, computedCeiling);

  const [amount, setAmount] = useState(issuable);
  const [reason, setReason] = useState("");
  const [failure, setFailure] = useState<string | null>(null);
  const [confirming, setConfirming] = useState<DecisionOutcome | null>(null);

  const reasonTooShort = reason.trim().length < MINIMUM_REASON;
  const aboveCap = Number(amount) > Number(issuable);
  const amountInvalid = !(Number(amount) > 0) || aboveCap;

  const submit = (outcome: DecisionOutcome) => {
    setFailure(null);
    setConfirming(null);

    startTransition(async () => {
      const result = await recordDecision(
        claimId,
        outcome === "approved"
          ? { outcome, approvedAmount: amount }
          : { outcome, reason },
        idempotencyKey,
      );

      if (!result.ok) {
        setFailure(result.message);
        return;
      }

      router.refresh();
    });
  };

  return (
    <div className="space-y-4 rounded-lg border border-neutral-200 bg-white px-4 py-4">
      <p className="font-medium">Record a decision</p>

      {failure ? (
        <div
          role="alert"
          className="rounded-md border border-danger-600 bg-danger-50 px-4 py-3"
        >
          <p className="text-caption text-danger-700">{failure}</p>
        </div>
      ) : null}

      {requiresDualApproval ? (
        <p className="rounded-md border border-info-600 bg-info-50 px-4 py-3 text-caption text-pretty text-info-700">
          This claim needs two approvals from different verifiers.{" "}
          {approvalsSoFar === 0
            ? "Yours would be the first."
            : "One approval is already recorded; yours would settle it."}
        </p>
      ) : null}

      <div className="space-y-2">
        <Label htmlFor="approved-amount">Approve for, tCO2e</Label>
        <Input
          id="approved-amount"
          inputMode="decimal"
          value={amount}
          onChange={(event) => setAmount(event.target.value)}
          aria-describedby="approved-amount-help"
        />
        <p id="approved-amount-help" className="text-caption text-neutral-600">
          At most {issuable} — the lower of the requested amount and the
          computed ceiling. This control cannot be set above it.
        </p>
        {aboveCap ? (
          <p className="text-caption text-danger-700" role="alert">
            {amount} exceeds {issuable}.
          </p>
        ) : null}
      </div>

      <div className="space-y-2">
        <Label htmlFor="decision-reason">
          Reason, for a rejection or an information request
        </Label>
        <Textarea
          id="decision-reason"
          rows={3}
          value={reason}
          onChange={(event) => setReason(event.target.value)}
          placeholder="Explain what is missing or wrong, in enough detail for the facility to act on."
        />
        <p className="text-caption text-neutral-600">
          {reason.trim().length} of {MINIMUM_REASON} characters minimum.
        </p>
      </div>

      <div className="flex flex-wrap gap-2 border-t border-neutral-200 pt-4">
        <Button
          type="button"
          disabled={pending || amountInvalid}
          onClick={() => setConfirming("approved")}
        >
          Approve
        </Button>
        <Button
          type="button"
          variant="outline"
          disabled={pending || reasonTooShort}
          onClick={() => setConfirming("more_information_requested")}
        >
          Request more information
        </Button>
        <Button
          type="button"
          variant="destructive"
          disabled={pending || reasonTooShort}
          onClick={() => setConfirming("rejected")}
        >
          Reject
        </Button>
      </div>

      <ConfirmDialog
        open={confirming !== null}
        onOpenChange={(open) => setConfirming(open ? confirming : null)}
        title={
          confirming === "approved"
            ? `Approve ${amount} tCO2e?`
            : confirming === "rejected"
              ? "Reject this claim?"
              : "Ask the facility for more information?"
        }
        description="A recorded decision is final. Correcting one afterwards is a separate, logged action taken by an admin, not an edit."
        confirmLabel="Record decision"
        onConfirm={() => confirming && submit(confirming)}
      />
    </div>
  );
}
