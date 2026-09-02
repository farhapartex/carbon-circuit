import { CreditAmountDisplay } from "@/components/shared/CreditAmountDisplay";
import { UsdcAmountDisplay } from "@/components/shared/UsdcAmountDisplay";
import type { CreditAmount, UsdcAmount } from "@/lib/decimal";

type PricePreviewProps = {
  quantity: CreditAmount | null;
  gross: UsdcAmount | null;
  fee: UsdcAmount | null;
  net: UsdcAmount | null;
  feeBasisPoints: number | null;
};

const percentOf = (basisPoints: number) =>
  `${(basisPoints / 100).toFixed(basisPoints % 100 === 0 ? 0 : 1)}%`;

export function PricePreview({
  quantity,
  gross,
  fee,
  net,
  feeBasisPoints,
}: PricePreviewProps) {
  return (
    <div className="rounded-lg border border-neutral-200 bg-neutral-50 px-4 py-4">
      <p className="font-medium">What you would receive</p>

      <dl className="mt-3 space-y-2">
        <div className="flex flex-wrap items-baseline justify-between gap-4">
          <dt className="text-caption text-neutral-600">Quantity listed</dt>
          <dd className="font-medium">
            {quantity === null ? (
              <span className="text-neutral-600">—</span>
            ) : (
              <CreditAmountDisplay amount={quantity} />
            )}
          </dd>
        </div>
        <div className="flex flex-wrap items-baseline justify-between gap-4">
          <dt className="text-caption text-neutral-600">Gross if fully sold</dt>
          <dd className="font-medium">
            {gross === null ? (
              <span className="text-neutral-600">—</span>
            ) : (
              <UsdcAmountDisplay amount={gross} exact />
            )}
          </dd>
        </div>
        <div className="flex flex-wrap items-baseline justify-between gap-4">
          <dt className="text-caption text-neutral-600">
            Platform fee
            {feeBasisPoints === null ? "" : ` (${percentOf(feeBasisPoints)})`}
          </dt>
          <dd className="font-medium">
            {fee === null ? (
              <span className="text-neutral-600">—</span>
            ) : (
              <>
                −<UsdcAmountDisplay amount={fee} exact />
              </>
            )}
          </dd>
        </div>
        <div className="flex flex-wrap items-baseline justify-between gap-4 border-t border-neutral-200 pt-2">
          <dt className="font-medium">Net proceeds</dt>
          <dd className="text-lg font-medium">
            {net === null ? (
              <span className="text-neutral-600">—</span>
            ) : (
              <UsdcAmountDisplay amount={net} exact />
            )}
          </dd>
        </div>
      </dl>

      {feeBasisPoints === null ? (
        <p className="mt-3 text-caption text-pretty text-neutral-600">
          Your plan does not permit selling, so no fee rate applies.
        </p>
      ) : (
        <p className="mt-3 text-caption text-pretty text-neutral-600">
          The fee is charged to you as the seller at your plan&apos;s rate.
          Buyers pay no platform fee.
        </p>
      )}
    </div>
  );
}
