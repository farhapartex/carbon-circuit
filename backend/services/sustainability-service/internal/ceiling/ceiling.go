package ceiling

import (
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

const Places = 6

var (
	ErrCapacityUnknown      = errors.New("no operating capacity is recorded for this facility")
	ErrPeriodOutsideVintage = errors.New("the claim period must fall inside its vintage year")
	ErrPeriodEmpty          = errors.New("the claim period must cover at least one day")
	ErrFactorUnknown        = errors.New("no reference factor applies to this claim")
	ErrDiscountUnknown      = errors.New("the facility carries no recognised verification discount")
)

var permittedDiscounts = map[string]struct{}{
	"1.00": {}, "0.75": {}, "0.50": {},
}

var thousand = decimal.NewFromInt(1000)

type Period struct {
	Start time.Time
	End   time.Time
}

func (p Period) days() int {
	return int(p.End.Sub(p.Start).Hours()/24) + 1
}

func (p Period) within(vintageYear int) bool {
	return p.Start.Year() == vintageYear && p.End.Year() == vintageYear
}

type Capacity struct {
	Value  decimal.Decimal
	Source string
}

const (
	Attested = "attested"
	Declared = "declared"
)

type Inputs struct {
	VintageYear    int
	Period         Period
	Capacity       Capacity
	ReferenceValue decimal.Decimal
	DiscountFactor decimal.Decimal
}

type Result struct {
	Ceiling      decimal.Decimal
	ProRated     decimal.Decimal
	PeriodDays   int
	VintageDays  int
	CapacityUsed decimal.Decimal
}

func daysInYear(year int) int {
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	return int(start.AddDate(1, 0, 0).Sub(start).Hours() / 24)
}

func (i Inputs) validate() error {
	if i.Capacity.Value.LessThanOrEqual(decimal.Zero) {
		return ErrCapacityUnknown
	}
	if i.ReferenceValue.LessThanOrEqual(decimal.Zero) {
		return ErrFactorUnknown
	}
	if _, known := permittedDiscounts[i.DiscountFactor.StringFixed(2)]; !known {
		return fmt.Errorf("%w: %s", ErrDiscountUnknown, i.DiscountFactor)
	}
	if i.Period.days() < 1 {
		return ErrPeriodEmpty
	}
	if !i.Period.within(i.VintageYear) {
		return fmt.Errorf("%w: period %s to %s, vintage %d",
			ErrPeriodOutsideVintage,
			i.Period.Start.Format(time.DateOnly),
			i.Period.End.Format(time.DateOnly),
			i.VintageYear)
	}
	return nil
}

func ForRenewableEnergy(inputs Inputs) (Result, error) {
	if err := inputs.validate(); err != nil {
		return Result{}, err
	}

	periodDays := inputs.Period.days()
	vintageDays := daysInYear(inputs.VintageYear)

	proRated := inputs.Capacity.Value.
		Mul(decimal.NewFromInt(int64(periodDays))).
		Div(decimal.NewFromInt(int64(vintageDays)))

	ceiling := proRated.
		Mul(inputs.ReferenceValue).
		Div(thousand).
		Mul(inputs.DiscountFactor).
		RoundDown(Places)

	return Result{
		Ceiling:      ceiling,
		ProRated:     proRated.RoundDown(Places),
		PeriodDays:   periodDays,
		VintageDays:  vintageDays,
		CapacityUsed: inputs.Capacity.Value,
	}, nil
}
