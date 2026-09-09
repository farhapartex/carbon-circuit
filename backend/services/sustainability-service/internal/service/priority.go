package service

import (
	"github.com/shopspring/decimal"

	"github.com/carboncircuit/backend/services/sustainability-service/internal/domain"
)

var dualApprovalThreshold = decimal.RequireFromString(domain.DualApprovalThreshold)

func IssuableAmount(requested, ceilingAmount decimal.Decimal) decimal.Decimal {
	if requested.LessThan(ceilingAmount) {
		return requested
	}
	return ceilingAmount
}

func RequiresDualApproval(requested, ceilingAmount decimal.Decimal) bool {
	return IssuableAmount(requested, ceilingAmount).GreaterThan(dualApprovalThreshold)
}

func PriorityFor(
	requested, ceilingAmount decimal.Decimal,
	discountFactor string,
) domain.QueuePriority {
	if RequiresDualApproval(requested, ceilingAmount) {
		return domain.Critical
	}
	if requested.GreaterThan(ceilingAmount) {
		return domain.High
	}
	if discountFactor == "0.50" {
		return domain.High
	}
	return domain.Normal
}
