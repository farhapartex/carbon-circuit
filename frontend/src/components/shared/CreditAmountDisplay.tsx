import { formatCreditAmount, type CreditAmount } from "@/lib/decimal";
import { cn } from "@/lib/utils";

export const CREDIT_UNIT = "tCO2e";

type CreditAmountDisplayProps = {
  amount: CreditAmount;
  className?: string;
  unitClassName?: string;
  showUnit?: boolean;
};

export function CreditAmountDisplay({
  amount,
  className,
  unitClassName,
  showUnit = true,
}: CreditAmountDisplayProps) {
  const formatted = formatCreditAmount(amount);

  return (
    <span
      data-slot="credit-amount"
      className={cn("tabular-nums", className)}
      title={`${formatted} ${CREDIT_UNIT}`}
    >
      {formatted}
      {showUnit ? (
        <span className={cn("ml-1 text-muted-foreground", unitClassName)}>
          {CREDIT_UNIT}
        </span>
      ) : null}
    </span>
  );
}
