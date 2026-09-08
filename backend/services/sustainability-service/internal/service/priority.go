package service

import (
	"github.com/shopspring/decimal"

	"github.com/carboncircuit/backend/services/sustainability-service/internal/domain"
)

var dualApprovalThreshold = decimal.RequireFromString(domain.DualApprovalThreshold)

func RequiresDualApproval(ceilingAmount decimal.Decimal) bool {
	return ceilingAmount.GreaterThan(dualApprovalThreshold)
}

func PriorityFor(
	requested, ceilingAmount decimal.Decimal,
	discountFactor string,
) domain.QueuePriority {
	if ceilingAmount.GreaterThan(dualApprovalThreshold) {
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
