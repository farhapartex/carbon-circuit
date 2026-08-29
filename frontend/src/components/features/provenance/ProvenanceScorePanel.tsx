import { ProvenanceScoreBadge } from "@/components/shared/StatusBadges";
import { provenanceBandForScore } from "@/lib/status";
import type { ProvenanceScore } from "@/lib/types";

const BAND_EXPLANATIONS = {
  complete: "Every expected step is recorded, anchored, and reported promptly.",
  strong:
    "Most of the journey is recorded and anchored, with minor gaps in the record.",
  partial:
    "Some steps are recorded, but this batch's journey is missing checkpoints or was reported late.",
  limited:
    "Very little of this batch's journey has been recorded so far. That means the record is thin, not that anything is wrong.",
} as const;

type ProvenanceScorePanelProps = {
  score: ProvenanceScore;
  showBreakdown?: boolean | undefined;
};

export function ProvenanceScorePanel({
  score,
  showBreakdown = true,
}: ProvenanceScorePanelProps) {
  const band = provenanceBandForScore(score.total);

  return (
    <section className="space-y-4 rounded-lg border border-neutral-200 bg-white p-6">
      <div className="flex flex-wrap items-center gap-3">
        <h2 className="font-medium">Provenance Score</h2>
        <ProvenanceScoreBadge score={score.total} />
      </div>
      <p className="text-caption text-pretty text-neutral-600">
        {BAND_EXPLANATIONS[band]}
      </p>

      {showBreakdown ? (
        <dl className="space-y-3 border-t border-neutral-200 pt-4">
          {score.components.map((component) => (
            <div key={component.label} className="space-y-1">
              <div className="flex items-baseline justify-between gap-4">
                <dt className="text-caption font-medium">{component.label}</dt>
                <dd className="text-caption text-neutral-600 tabular-nums">
                  {component.earned} / {component.available}
                </dd>
              </div>
              <progress
                data-slot="score-meter"
                value={component.earned}
                max={component.available}
                aria-label={`${component.label}: ${component.earned} of ${component.available} points`}
              />
              <p className="text-caption text-neutral-600">
                {component.explanation}
              </p>
            </div>
          ))}
        </dl>
      ) : null}
    </section>
  );
}
