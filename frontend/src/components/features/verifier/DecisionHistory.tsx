import { CheckCircle2, CircleHelp, XCircle } from "lucide-react";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { DecisionOutcome, DecisionRecord } from "@/lib/api/verifier";

const presentation: Record<
  DecisionOutcome,
  { icon: typeof CheckCircle2; label: string; className: string }
> = {
  approved: {
    icon: CheckCircle2,
    label: "Approved",
    className: "text-success-700",
  },
  rejected: { icon: XCircle, label: "Rejected", className: "text-danger-700" },
  more_information_requested: {
    icon: CircleHelp,
    label: "More information requested",
    className: "text-info-700",
  },
};

export function DecisionHistory({
  decisions,
  requiresDualApproval,
}: {
  decisions: DecisionRecord[];
  requiresDualApproval: boolean;
}) {
  const approvals = decisions.filter(
    (decision) => decision.outcome === "approved",
  );

  const awaitingSecond = requiresDualApproval && approvals.length === 1;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Decisions</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {awaitingSecond ? (
          <div className="rounded-md border border-info-600 bg-info-50 px-4 py-3">
            <p className="font-medium text-info-700">
              One approval recorded, a second is required
            </p>
            <p className="mt-1 text-caption text-pretty text-info-700">
              This claim would issue more than 5,000 tCO2e, so a second verifier
              must agree before anything is issued. It cannot be the same
              person, and the amount issued is the lower of the two.
            </p>
          </div>
        ) : null}

        {decisions.length === 0 ? (
          <p className="text-caption text-neutral-600">
            No decision has been recorded yet.
          </p>
        ) : (
          decisions.map((decision) => {
            const shown = presentation[decision.outcome];
            const Icon = shown.icon;

            return (
              <div
                key={decision.id}
                className="rounded-md border border-neutral-200 px-4 py-3"
              >
                <div className="flex flex-wrap items-baseline justify-between gap-2">
                  <span
                    className={`flex items-center gap-1.5 font-medium ${shown.className}`}
                  >
                    <Icon className="size-4 shrink-0" aria-hidden />
                    {shown.label}
                    {decision.approvedAmount
                      ? ` at ${decision.approvedAmount} tCO2e`
                      : ""}
                  </span>
                  <span className="text-caption text-neutral-600">
                    {decision.verifierName} ·{" "}
                    <TimestampDisplay value={decision.decidedAt} />
                  </span>
                </div>
                {decision.reason ? (
                  <p className="mt-2 text-caption text-pretty text-neutral-600">
                    {decision.reason}
                  </p>
                ) : null}
              </div>
            );
          })
        )}
      </CardContent>
    </Card>
  );
}
