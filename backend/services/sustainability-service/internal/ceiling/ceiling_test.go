package ceiling_test

import (
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/carboncircuit/backend/services/sustainability-service/internal/ceiling"
)

func amount(t *testing.T, value string) decimal.Decimal {
	t.Helper()

	parsed, err := decimal.NewFromString(value)
	if err != nil {
		t.Fatalf("parse %q: %v", value, err)
	}
	return parsed
}

func day(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		t.Fatalf("parse %q: %v", value, err)
	}
	return parsed
}

func fullYear(t *testing.T, year int) ceiling.Period {
	t.Helper()

	return ceiling.Period{
		Start: day(t, "2026-01-01"),
		End:   day(t, "2026-12-31"),
	}
}

func taiwanInputs(t *testing.T) ceiling.Inputs {
	t.Helper()

	return ceiling.Inputs{
		VintageYear:    2026,
		Period:         fullYear(t, 2026),
		Capacity:       ceiling.Capacity{Value: amount(t, "12400000"), Source: ceiling.Attested},
		ReferenceValue: amount(t, "0.494"),
		DiscountFactor: amount(t, "1.00"),
	}
}

func TestTheWorkedExampleFromThePRD(t *testing.T) {
	result, err := ceiling.ForRenewableEnergy(taiwanInputs(t))
	if err != nil {
		t.Fatalf("compute: %v", err)
	}

	if got := result.Ceiling.String(); got != "6125.6" {
		t.Fatalf("PRD 4.2 states a ceiling of 6125.6 tCO2e, got %s", got)
	}
}

func TestTheRequestedFigureCannotRaiseTheCeiling(t *testing.T) {
	inputs := taiwanInputs(t)

	modest, err := ceiling.ForRenewableEnergy(inputs)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}

	if !modest.CapacityUsed.Equal(amount(t, "12400000")) {
		t.Fatal("the ceiling must be computed from capacity, and nothing else is an input")
	}

	if got := modest.Ceiling.String(); got != "6125.6" {
		t.Fatalf("expected the capacity-derived ceiling, got %s", got)
	}
}

func TestEachDiscountFactorScalesTheCeiling(t *testing.T) {
	cases := map[string]string{
		"1.00": "6125.6",
		"0.75": "4594.2",
		"0.50": "3062.8",
	}

	for discount, expected := range cases {
		t.Run(discount, func(t *testing.T) {
			inputs := taiwanInputs(t)
			inputs.DiscountFactor = amount(t, discount)

			result, err := ceiling.ForRenewableEnergy(inputs)
			if err != nil {
				t.Fatalf("compute: %v", err)
			}
			if got := result.Ceiling.String(); got != expected {
				t.Fatalf("discount %s should give %s, got %s", discount, expected, got)
			}
		})
	}
}

func TestASelfDeclaredFacilityCannotClaimMoreThanHalf(t *testing.T) {
	corroborated := taiwanInputs(t)

	selfDeclared := taiwanInputs(t)
	selfDeclared.DiscountFactor = amount(t, "0.50")
	selfDeclared.Capacity.Source = ceiling.Declared

	full, err := ceiling.ForRenewableEnergy(corroborated)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	half, err := ceiling.ForRenewableEnergy(selfDeclared)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}

	if !half.Ceiling.Mul(decimal.NewFromInt(2)).Equal(full.Ceiling) {
		t.Fatalf("PRD 2.1 requires exactly half for an uncorroborated facility: %s vs %s",
			half.Ceiling, full.Ceiling)
	}
}

func TestAQuarterOfTheYearEarnsAQuarterOfTheCapacity(t *testing.T) {
	inputs := taiwanInputs(t)
	inputs.Period = ceiling.Period{Start: day(t, "2026-01-01"), End: day(t, "2026-03-31")}

	result, err := ceiling.ForRenewableEnergy(inputs)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}

	if result.PeriodDays != 90 {
		t.Fatalf("expected 90 days in Q1 2026, got %d", result.PeriodDays)
	}
	if result.VintageDays != 365 {
		t.Fatalf("expected 365 days in 2026, got %d", result.VintageDays)
	}

	expected := amount(t, "12400000").
		Mul(decimal.NewFromInt(90)).
		Div(decimal.NewFromInt(365)).
		Mul(amount(t, "0.494")).
		Div(decimal.NewFromInt(1000)).
		RoundDown(6)

	if !result.Ceiling.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected, result.Ceiling)
	}
	if !result.Ceiling.LessThan(amount(t, "6125.6")) {
		t.Fatal("a quarterly claim must not reach the full annual ceiling")
	}
}

func TestFourQuartersCannotExceedTheAnnualCeiling(t *testing.T) {
	quarters := []ceiling.Period{
		{Start: day(t, "2026-01-01"), End: day(t, "2026-03-31")},
		{Start: day(t, "2026-04-01"), End: day(t, "2026-06-30")},
		{Start: day(t, "2026-07-01"), End: day(t, "2026-09-30")},
		{Start: day(t, "2026-10-01"), End: day(t, "2026-12-31")},
	}

	total := decimal.Zero
	days := 0

	for _, quarter := range quarters {
		inputs := taiwanInputs(t)
		inputs.Period = quarter

		result, err := ceiling.ForRenewableEnergy(inputs)
		if err != nil {
			t.Fatalf("compute: %v", err)
		}
		total = total.Add(result.Ceiling)
		days += result.PeriodDays
	}

	if days != 365 {
		t.Fatalf("the four quarters should cover the year exactly, got %d days", days)
	}

	annual, err := ceiling.ForRenewableEnergy(taiwanInputs(t))
	if err != nil {
		t.Fatalf("compute: %v", err)
	}

	if total.GreaterThan(annual.Ceiling) {
		t.Fatalf("four quarterly ceilings (%s) must not exceed the annual ceiling (%s)",
			total, annual.Ceiling)
	}
}

func TestALeapYearUses366Days(t *testing.T) {
	inputs := taiwanInputs(t)
	inputs.VintageYear = 2028
	inputs.Period = ceiling.Period{Start: day(t, "2028-01-01"), End: day(t, "2028-12-31")}

	result, err := ceiling.ForRenewableEnergy(inputs)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}

	if result.VintageDays != 366 {
		t.Fatalf("2028 is a leap year, expected 366 days, got %d", result.VintageDays)
	}
	if result.PeriodDays != 366 {
		t.Fatalf("expected the full leap year, got %d", result.PeriodDays)
	}
}

func TestAPeriodOutsideItsVintageIsRefused(t *testing.T) {
	inputs := taiwanInputs(t)
	inputs.Period = ceiling.Period{Start: day(t, "2025-07-01"), End: day(t, "2026-06-30")}

	_, err := ceiling.ForRenewableEnergy(inputs)
	if !errors.Is(err, ceiling.ErrPeriodOutsideVintage) {
		t.Fatalf("a period spanning two years would pro-rate past annual capacity, got %v", err)
	}
}

func TestATwoYearPeriodCannotDoubleTheCapacity(t *testing.T) {
	inputs := taiwanInputs(t)
	inputs.Period = ceiling.Period{Start: day(t, "2026-01-01"), End: day(t, "2027-12-31")}

	_, err := ceiling.ForRenewableEnergy(inputs)
	if err == nil {
		t.Fatal("a two year period must be refused rather than pro-rated to 200% of capacity")
	}
}

func TestAnUnknownDiscountIsRefused(t *testing.T) {
	for _, discount := range []string{"0.00", "0.90", "1.50", "-1.00"} {
		t.Run(discount, func(t *testing.T) {
			inputs := taiwanInputs(t)
			inputs.DiscountFactor = amount(t, discount)

			_, err := ceiling.ForRenewableEnergy(inputs)
			if !errors.Is(err, ceiling.ErrDiscountUnknown) {
				t.Fatalf("expected ErrDiscountUnknown, got %v", err)
			}
		})
	}
}

func TestMissingCapacityIsRefusedRatherThanTreatedAsZero(t *testing.T) {
	inputs := taiwanInputs(t)
	inputs.Capacity.Value = decimal.Zero

	_, err := ceiling.ForRenewableEnergy(inputs)
	if !errors.Is(err, ceiling.ErrCapacityUnknown) {
		t.Fatalf("expected ErrCapacityUnknown, got %v", err)
	}
}

func TestMissingReferenceFactorIsRefused(t *testing.T) {
	inputs := taiwanInputs(t)
	inputs.ReferenceValue = decimal.Zero

	_, err := ceiling.ForRenewableEnergy(inputs)
	if !errors.Is(err, ceiling.ErrFactorUnknown) {
		t.Fatalf("expected ErrFactorUnknown, got %v", err)
	}
}

func TestASingleDayPeriodIsAllowed(t *testing.T) {
	inputs := taiwanInputs(t)
	inputs.Period = ceiling.Period{Start: day(t, "2026-06-01"), End: day(t, "2026-06-01")}

	result, err := ceiling.ForRenewableEnergy(inputs)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	if result.PeriodDays != 1 {
		t.Fatalf("expected one day, got %d", result.PeriodDays)
	}
	if !result.Ceiling.GreaterThan(decimal.Zero) {
		t.Fatal("a one day claim should still earn a positive ceiling")
	}
}

func TestTheCeilingRoundsDownNeverUp(t *testing.T) {
	inputs := taiwanInputs(t)
	inputs.Period = ceiling.Period{Start: day(t, "2026-01-01"), End: day(t, "2026-01-07")}

	result, err := ceiling.ForRenewableEnergy(inputs)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}

	exact := amount(t, "12400000").
		Mul(decimal.NewFromInt(7)).
		Div(decimal.NewFromInt(365)).
		Mul(amount(t, "0.494")).
		Div(decimal.NewFromInt(1000))

	if result.Ceiling.GreaterThan(exact) {
		t.Fatalf("a cap must never round upward: %s exceeds the exact %s",
			result.Ceiling, exact)
	}
	if result.Ceiling.Exponent() < -6 {
		t.Fatalf("expected at most 6 decimal places, got %s", result.Ceiling)
	}
}

func TestEveryGridFactorProducesAPositiveCeiling(t *testing.T) {
	factors := []string{
		"0.237", "0.396", "0.324", "0.438", "0.381", "0.056", "0.662",
		"0.207", "0.581", "0.512", "0.713", "0.462", "0.436", "0.494",
		"0.521", "0.585", "0.412", "0.499",
	}

	for _, factor := range factors {
		inputs := taiwanInputs(t)
		inputs.ReferenceValue = amount(t, factor)

		result, err := ceiling.ForRenewableEnergy(inputs)
		if err != nil {
			t.Fatalf("factor %s: %v", factor, err)
		}
		if !result.Ceiling.GreaterThan(decimal.Zero) {
			t.Fatalf("factor %s produced a non-positive ceiling", factor)
		}
	}
}

func TestAFreshVintageOffersItsWholeCeiling(t *testing.T) {
	allowance, err := ceiling.Allow(taiwanInputs(t), decimal.Zero)
	if err != nil {
		t.Fatalf("allow: %v", err)
	}

	if got := allowance.Effective.String(); got != "6125.6" {
		t.Fatalf("expected the full ceiling, got %s", got)
	}
	if !allowance.Remaining.Equal(allowance.VintageCeiling) {
		t.Fatal("nothing consumed, so remaining should equal the vintage ceiling")
	}
}

func TestEarlierClaimsReduceWhatIsLeft(t *testing.T) {
	allowance, err := ceiling.Allow(taiwanInputs(t), amount(t, "4000"))
	if err != nil {
		t.Fatalf("allow: %v", err)
	}

	if got := allowance.Remaining.String(); got != "2125.6" {
		t.Fatalf("expected 6125.6 less 4000, got %s", got)
	}
	if got := allowance.Effective.String(); got != "2125.6" {
		t.Fatalf("the effective ceiling must fall to what remains, got %s", got)
	}
}

func TestAnExhaustedVintageIsRefused(t *testing.T) {
	_, err := ceiling.Allow(taiwanInputs(t), amount(t, "6125.6"))
	if !errors.Is(err, ceiling.ErrVintageExhausted) {
		t.Fatalf("expected ErrVintageExhausted, got %v", err)
	}
}

func TestOverConsumptionCannotProduceANegativeAllowance(t *testing.T) {
	allowance, err := ceiling.Allow(taiwanInputs(t), amount(t, "99999"))
	if !errors.Is(err, ceiling.ErrVintageExhausted) {
		t.Fatalf("expected ErrVintageExhausted, got %v", err)
	}
	if allowance.Remaining.IsNegative() || allowance.Effective.IsNegative() {
		t.Fatalf("an over-consumed vintage must clamp at zero, got remaining %s effective %s",
			allowance.Remaining, allowance.Effective)
	}
}

func TestAShortPeriodStillCannotExceedItsOwnShare(t *testing.T) {
	inputs := taiwanInputs(t)
	inputs.Period = ceiling.Period{Start: day(t, "2026-01-01"), End: day(t, "2026-03-31")}

	allowance, err := ceiling.Allow(inputs, decimal.Zero)
	if err != nil {
		t.Fatalf("allow: %v", err)
	}

	if !allowance.Effective.Equal(allowance.PeriodCeiling) {
		t.Fatal("with the whole vintage free, a quarter should earn exactly its pro-rated share")
	}
	if !allowance.Effective.LessThan(allowance.VintageCeiling) {
		t.Fatal("a quarterly claim must not reach the annual ceiling")
	}
}

func TestFourQuartersConsumeTheVintageExactlyOnce(t *testing.T) {
	quarters := []ceiling.Period{
		{Start: day(t, "2026-01-01"), End: day(t, "2026-03-31")},
		{Start: day(t, "2026-04-01"), End: day(t, "2026-06-30")},
		{Start: day(t, "2026-07-01"), End: day(t, "2026-09-30")},
		{Start: day(t, "2026-10-01"), End: day(t, "2026-12-31")},
	}

	consumed := decimal.Zero

	for index, quarter := range quarters {
		inputs := taiwanInputs(t)
		inputs.Period = quarter

		allowance, err := ceiling.Allow(inputs, consumed)
		if err != nil {
			t.Fatalf("quarter %d: %v", index+1, err)
		}

		consumed = consumed.Add(allowance.Effective)
	}

	annual, err := ceiling.Allow(taiwanInputs(t), decimal.Zero)
	if err != nil {
		t.Fatalf("annual: %v", err)
	}

	if consumed.GreaterThan(annual.VintageCeiling) {
		t.Fatalf("four quarters consumed %s against a vintage ceiling of %s",
			consumed, annual.VintageCeiling)
	}
}

func TestASecondFullYearClaimGetsNothing(t *testing.T) {
	first, err := ceiling.Allow(taiwanInputs(t), decimal.Zero)
	if err != nil {
		t.Fatalf("first: %v", err)
	}

	_, err = ceiling.Allow(taiwanInputs(t), first.Effective)
	if !errors.Is(err, ceiling.ErrVintageExhausted) {
		t.Fatalf("a facility cannot claim its whole year twice, got %v", err)
	}
}
