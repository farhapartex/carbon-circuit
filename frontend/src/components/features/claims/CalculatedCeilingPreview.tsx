"use client";

import { Calculator, Loader2 } from "lucide-react";
import { useEffect, useState, useTransition } from "react";
import { loadCeilingPreview } from "@/lib/actions/claims";
import type { CeilingPreview } from "@/lib/api/claims";
import type { ActivityType } from "@/lib/types";

type Answer = {
  signature: string;
  preview: CeilingPreview | null;
  failure: string | null;
};

type CalculatedCeilingPreviewProps = {
  facilityId: string;
  activityType: ActivityType;
  vintageYear: string;
  periodStart: string;
  periodEnd: string;
};

const capacityWording: Record<string, string> = {
  attested: "independently attested",
  declared: "self-declared",
};

export function CalculatedCeilingPreview({
  facilityId,
  activityType,
  vintageYear,
  periodStart,
  periodEnd,
}: CalculatedCeilingPreviewProps) {
  const [pending, startTransition] = useTransition();
  const [answered, setAnswered] = useState<Answer | null>(null);

  const ready =
    facilityId !== "" &&
    /^\d{4}$/.test(vintageYear) &&
    periodStart !== "" &&
    periodEnd !== "";

  const signature = [
    facilityId,
    activityType,
    vintageYear,
    periodStart,
    periodEnd,
  ].join("|");

  useEffect(() => {
    if (!ready) return;

    startTransition(async () => {
      const result = await loadCeilingPreview({
        facilityId,
        activityType,
        vintageYear: Number(vintageYear),
        periodStart,
        periodEnd,
      });

      setAnswered({
        signature,
        preview: result.ok ? result.preview : null,
        failure: result.ok ? null : result.message,
      });
    });
  }, [
    ready,
    signature,
    facilityId,
    activityType,
    vintageYear,
    periodStart,
    periodEnd,
  ]);

  const current = answered?.signature === signature ? answered : null;
  const preview = current?.preview ?? null;
  const failure = current?.failure ?? null;

  return (
    <div className="rounded-lg border border-neutral-200 bg-neutral-50 px-4 py-4">
      <p className="flex items-center gap-2 font-medium">
        <Calculator className="size-4 shrink-0 text-neutral-600" aria-hidden />
        Your credit ceiling
        {pending ? (
          <Loader2
            className="size-3.5 animate-spin text-neutral-600"
            aria-hidden
          />
        ) : null}
      </p>

      {!ready ? (
        <p className="mt-2 text-caption text-pretty text-neutral-600">
          Choose a facility and a claim period and the server will compute your
          ceiling from that facility&apos;s operating capacity.
        </p>
      ) : null}

      {failure ? (
        <p
          className="mt-2 text-caption text-pretty text-danger-700"
          role="alert"
        >
          {failure}
        </p>
      ) : null}

      {preview ? (
        <>
          <p className="text-h3 mt-3 tabular-nums">
            {preview.ceiling}{" "}
            <span className="text-caption text-neutral-600">tCO2e</span>
          </p>
          <dl className="mt-4 space-y-2 border-t border-neutral-200 pt-4">
            <div className="flex flex-wrap items-baseline justify-between gap-4">
              <dt className="text-caption text-neutral-600">
                Capacity it is computed from
              </dt>
              <dd className="font-medium tabular-nums">
                {preview.capacityBasis} kWh
                <span className="ml-1 text-caption font-normal text-neutral-600">
                  (
                  {capacityWording[preview.capacitySource] ??
                    preview.capacitySource}
                  )
                </span>
              </dd>
            </div>
            <div className="flex flex-wrap items-baseline justify-between gap-4">
              <dt className="text-caption text-neutral-600">
                Grid factor, {preview.gridRegion}
              </dt>
              <dd className="font-medium tabular-nums">
                {preview.referenceValue} kgCO2e/kWh
              </dd>
            </div>
            <div className="flex flex-wrap items-baseline justify-between gap-4">
              <dt className="text-caption text-neutral-600">
                Verification discount
              </dt>
              <dd className="font-medium tabular-nums">
                ×{preview.discountFactor}
              </dd>
            </div>
            <div className="flex flex-wrap items-baseline justify-between gap-4">
              <dt className="text-caption text-neutral-600">Period</dt>
              <dd className="font-medium tabular-nums">
                {preview.periodDays} of {preview.vintageDays} days
              </dd>
            </div>
          </dl>

          {Number(preview.consumed) > 0 ? (
            <dl className="mt-4 space-y-2 border-t border-neutral-200 pt-4">
              <div className="flex flex-wrap items-baseline justify-between gap-4">
                <dt className="text-caption text-neutral-600">
                  Full year for this vintage
                </dt>
                <dd className="font-medium tabular-nums">
                  {preview.vintageCeiling}
                </dd>
              </div>
              <div className="flex flex-wrap items-baseline justify-between gap-4">
                <dt className="text-caption text-neutral-600">
                  Already claimed or issued
                </dt>
                <dd className="font-medium text-warning-700 tabular-nums">
                  −{preview.consumed}
                </dd>
              </div>
              <div className="flex flex-wrap items-baseline justify-between gap-4">
                <dt className="text-caption text-neutral-600">
                  Left for this vintage
                </dt>
                <dd className="font-medium tabular-nums">
                  {preview.remaining}
                </dd>
              </div>
            </dl>
          ) : null}
          <p className="mt-3 text-caption text-pretty text-neutral-600">
            The amount you request cannot raise this figure. Nothing can be
            issued above it, whatever a verifier approves.
            {Number(preview.consumed) > 0
              ? " Earlier claims for this facility and vintage draw from the same annual capacity."
              : ""}
          </p>
        </>
      ) : null}
    </div>
  );
}
