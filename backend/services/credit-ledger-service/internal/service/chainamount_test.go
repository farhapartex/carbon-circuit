package service_test

import (
	"math/big"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/carboncircuit/backend/services/credit-ledger-service/internal/service"
)

func figure(t *testing.T, value string) decimal.Decimal {
	t.Helper()

	parsed, err := decimal.NewFromString(value)
	if err != nil {
		t.Fatalf("parse %q: %v", value, err)
	}
	return parsed
}

func TestSixDecimalPlacesBecomeEighteenDecimalIntegers(t *testing.T) {
	cases := map[string]string{
		"1":           "1000000000000000000",
		"400":         "400000000000000000000",
		"400.000000":  "400000000000000000000",
		"0.000001":    "1000000000000",
		"6125.6":      "6125600000000000000000",
		"5639.998":    "5639998000000000000000",
		"1390.684438": "1390684438000000000000",
	}

	for amount, expected := range cases {
		t.Run(amount, func(t *testing.T) {
			if got := service.ChainAmount(figure(t, amount)); got != expected {
				t.Fatalf("expected %s wei, got %s", expected, got)
			}
		})
	}
}

func TestTheConversionLosesNothing(t *testing.T) {
	amount := figure(t, "1390.684438")

	wei, ok := new(big.Int).SetString(service.ChainAmount(amount), 10)
	if !ok {
		t.Fatal("conversion did not produce an integer")
	}

	back := decimal.NewFromBigInt(wei, -18)

	if !back.Equal(amount) {
		t.Fatalf("round trip changed %s into %s", amount, back)
	}
}

func TestTheLargestPlausibleIssuanceStillFits(t *testing.T) {
	amount := figure(t, "9999999999999999999999.999999")

	wei, ok := new(big.Int).SetString(service.ChainAmount(amount), 10)
	if !ok {
		t.Fatal("conversion did not produce an integer")
	}

	if wei.BitLen() > 256 {
		t.Fatalf("a numeric(28,6) maximum needs %d bits, more than uint256 holds", wei.BitLen())
	}
}

func TestMorePrecisionThanTheLedgerHoldsIsRefused(t *testing.T) {
	tooFine := decimal.New(1, -19)

	if got := service.ChainAmount(tooFine); got != "" {
		t.Fatalf("an amount finer than a wei must be refused, got %s", got)
	}
}
