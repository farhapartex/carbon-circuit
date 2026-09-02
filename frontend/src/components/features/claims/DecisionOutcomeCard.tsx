import { CreditAmountDisplay } from "@/components/shared/CreditAmountDisplay";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { StatusPill } from "@/components/shared/StatusPill";
import type { StatusVariant } from "@/lib/status";
import type { SustainabilityClaim, VerifierDecisionOutcome } from "@/lib/types";

const outcomePresentation: Record<
  VerifierDecisionOutcome,
  { label: string; variant: StatusVariant }
> = {
  approved: { label: "Approved", variant: "success" },
  rejected: { label: "Rejected", variant: "danger" },
  more_information_requested: {
    label: "More information requested",
    variant: "info",
  },
};

export function DecisionOutcomeCard({ claim }: { claim: SustainabilityClaim }) {
  if (claim.decisions.length === 0) return null;

  return (
    <Card>
      <CardHeader>
        <CardTitle>
          {claim.decisions.length === 1 ? "Decision" : "Decisions"}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {claim.requiresDualApproval ? (
          <p className="rounded-md border border-neutral-200 bg-neutral-50 px-4 py-3 text-caption text-pretty text-neutral-600">
            This claim exceeds the dual-approval threshold, so it needs two
            independent verifier approvals before credits are issued.
          </p>
        ) : null}

        {claim.decisions.map((decision) => (
          <div
            key={decision.id}
            className="space-y-2 rounded-md border border-neutral-200 px-4 py-3"
          >
            <div className="flex flex-wrap items-center justify-between gap-2">
              <StatusPill
                presentation={outcomePresentation[decision.outcome]}
              />
              <span className="text-caption text-neutral-600">
                <TimestampDisplay value={decision.decidedAt} />
              </span>
            </div>

            {decision.approvedAmount ? (
              <p className="flex flex-wrap items-baseline justify-between gap-4">
                <span className="text-caption text-neutral-600">
                  Approved amount
                </span>
                <span className="font-medium">
                  <CreditAmountDisplay amount={decision.approvedAmount} />
                </span>
              </p>
            ) : null}

            <p className="text-neutral-700 text-caption text-pretty">
              {decision.reason}
            </p>
            <p className="text-caption text-neutral-600">
              {decision.verifierName}
            </p>
          </div>
        ))}

        {claim.issuedAmount ? (
          <p className="flex flex-wrap items-baseline justify-between gap-4 border-t border-neutral-200 pt-4">
            <span className="text-caption text-neutral-600">
              Credits issued
            </span>
            <span className="text-lg font-medium">
              <CreditAmountDisplay amount={claim.issuedAmount} />
            </span>
          </p>
        ) : null}
      </CardContent>
    </Card>
  );
}
