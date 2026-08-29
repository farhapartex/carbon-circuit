import * as dn from "dnum";

declare const creditAmountBrand: unique symbol;
declare const usdcAmountBrand: unique symbol;

export type CreditAmount = string & { readonly [creditAmountBrand]: true };
export type UsdcAmount = string & { readonly [usdcAmountBrand]: true };

export const CREDIT_DECIMALS = 6;
export const USDC_DECIMALS = 6;
export const USDC_DISPLAY_DECIMALS = 2;

const PLAIN_DECIMAL = /^-?(0|[1-9]\d*)(\.\d+)?$/;

const fractionDigitCount = (value: string) => value.split(".")[1]?.length ?? 0;

const parsePlainDecimal = (
  value: string,
  maximumDecimals: number,
  kind: string,
): string => {
  if (!PLAIN_DECIMAL.test(value)) {
    throw new Error(
      `${kind} must be a plain decimal string, received "${value}"`,
    );
  }
  if (fractionDigitCount(value) > maximumDecimals) {
    throw new Error(
      `${kind} carries more than ${maximumDecimals} decimal places, received "${value}"`,
    );
  }
  return value;
};

const canonicalise = (value: string, decimals: number) =>
  dn.toString(dn.from(value, decimals));

export const creditAmount = (value: string): CreditAmount =>
  canonicalise(
    parsePlainDecimal(value, CREDIT_DECIMALS, "Credit amount"),
    CREDIT_DECIMALS,
  ) as CreditAmount;

export const usdcAmount = (value: string): UsdcAmount =>
  canonicalise(
    parsePlainDecimal(value, USDC_DECIMALS, "USDC amount"),
    USDC_DECIMALS,
  ) as UsdcAmount;

const asCredits = (amount: CreditAmount) => dn.from(amount, CREDIT_DECIMALS);
const asUsdc = (amount: UsdcAmount) => dn.from(amount, USDC_DECIMALS);

export const formatCreditAmount = (amount: CreditAmount): string =>
  dn.format(asCredits(amount), { digits: CREDIT_DECIMALS });

export const formatUsdcAmount = (
  amount: UsdcAmount,
  options?: { exact?: boolean },
): string =>
  options?.exact === true
    ? dn.format(asUsdc(amount), { digits: USDC_DECIMALS })
    : dn.format(asUsdc(amount), {
        digits: USDC_DISPLAY_DECIMALS,
        trailingZeros: true,
      });

export const addCreditAmounts = (
  a: CreditAmount,
  b: CreditAmount,
): CreditAmount =>
  dn.toString(dn.add(asCredits(a), asCredits(b))) as CreditAmount;

export const subtractCreditAmounts = (
  a: CreditAmount,
  b: CreditAmount,
): CreditAmount =>
  dn.toString(dn.subtract(asCredits(a), asCredits(b))) as CreditAmount;

export const compareCreditAmounts = (
  a: CreditAmount,
  b: CreditAmount,
): number => dn.compare(asCredits(a), asCredits(b));

export const compareUsdcAmounts = (a: UsdcAmount, b: UsdcAmount): number =>
  dn.compare(asUsdc(a), asUsdc(b));

export const isZeroCreditAmount = (amount: CreditAmount): boolean =>
  dn.equal(asCredits(amount), dn.from(0, CREDIT_DECIMALS));

export const creditAmountToBaseUnits = (amount: CreditAmount): bigint =>
  asCredits(amount)[0];

export const creditAmountFromBaseUnits = (baseUnits: bigint): CreditAmount =>
  dn.toString([baseUnits, CREDIT_DECIMALS]) as CreditAmount;

export const usdcAmountToBaseUnits = (amount: UsdcAmount): bigint =>
  asUsdc(amount)[0];

export const usdcAmountFromBaseUnits = (baseUnits: bigint): UsdcAmount =>
  dn.toString([baseUnits, USDC_DECIMALS]) as UsdcAmount;
