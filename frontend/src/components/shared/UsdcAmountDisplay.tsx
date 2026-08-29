import { formatUsdcAmount, type UsdcAmount } from "@/lib/decimal";
import { cn } from "@/lib/utils";

export const USDC_UNIT = "USDC";

type UsdcAmountDisplayProps = {
  amount: UsdcAmount;
  className?: string;
  unitClassName?: string;
  showUnit?: boolean;
  exact?: boolean;
};

export function UsdcAmountDisplay({
  amount,
  className,
  unitClassName,
  showUnit = true,
  exact = false,
}: UsdcAmountDisplayProps) {
  const formatted = formatUsdcAmount(amount, { exact });
  const settled = formatUsdcAmount(amount, { exact: true });

  return (
    <span
      data-slot="usdc-amount"
      className={cn("tabular-nums", className)}
      title={`${settled} ${USDC_UNIT}`}
    >
      {formatted}
      {showUnit ? (
        <span className={cn("ml-1 text-muted-foreground", unitClassName)}>
          {USDC_UNIT}
        </span>
      ) : null}
    </span>
  );
}
