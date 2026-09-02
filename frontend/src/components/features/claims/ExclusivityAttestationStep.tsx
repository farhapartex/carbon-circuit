"use client";

import { ScrollText } from "lucide-react";

type ExclusivityAttestationStepProps = {
  attested: boolean;
  onAttestedChange: (attested: boolean) => void;
  userName: string;
};

export function ExclusivityAttestationStep({
  attested,
  onAttestedChange,
  userName,
}: ExclusivityAttestationStepProps) {
  return (
    <div className="space-y-4">
      <div className="border-neutral-300 rounded-lg border bg-white px-5 py-5">
        <p className="flex items-center gap-2 font-medium">
          <ScrollText
            className="size-4 shrink-0 text-primary-600"
            aria-hidden
          />
          Exclusivity attestation
        </p>

        <p className="mt-3 text-body text-pretty">
          I confirm that the emissions reduction described in this claim has not
          been, and will not be, registered with any other carbon registry,
          offset programme, or crediting scheme.
        </p>

        <p className="mt-3 text-caption text-pretty text-neutral-600">
          This is a legal representation, not a preference. It is recorded with
          a timestamp and tied to your user account permanently, and it stays
          attached to the claim after submission rather than disappearing with
          the form. Claiming the same reduction in two places is the specific
          failure this platform exists to prevent.
        </p>

        <label className="mt-5 flex items-start gap-3 border-t border-neutral-200 pt-4">
          <input
            type="checkbox"
            checked={attested}
            onChange={(event) => onAttestedChange(event.target.checked)}
            className="mt-0.5"
          />
          <span className="text-body text-pretty">
            I affirm the statement above, as {userName}.
          </span>
        </label>
      </div>
    </div>
  );
}
