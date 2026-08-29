import { Leaf } from "lucide-react";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import type { PublicClaimSummary } from "@/lib/types";

type SustainabilityHighlightCardProps = {
  facilityName: string;
  claims: PublicClaimSummary[];
};

export function SustainabilityHighlightCard({
  facilityName,
  claims,
}: SustainabilityHighlightCardProps) {
  if (claims.length === 0) return null;

  return (
    <section className="space-y-3 rounded-lg border border-primary-600/30 bg-primary-50 p-6">
      <h2 className="inline-flex items-center gap-2 font-medium text-primary-800">
        <Leaf className="size-4" aria-hidden />
        Verified sustainability practice
      </h2>
      <p className="text-caption text-pretty text-primary-800/80">
        {facilityName} has had the practices below independently reviewed and
        approved by a human verifier.
      </p>
      <ul className="space-y-2">
        {claims.map((claim) => (
          <li
            key={`${claim.activityTypeLabel}-${claim.vintageYear}`}
            className="flex flex-wrap items-baseline gap-x-2 text-caption text-primary-800"
          >
            <span className="font-medium">{claim.activityTypeLabel}</span>
            <span>· vintage {claim.vintageYear}</span>
            <span className="text-primary-800/70">
              · approved <TimestampDisplay value={claim.approvedAt} dateOnly />
            </span>
          </li>
        ))}
      </ul>
    </section>
  );
}
