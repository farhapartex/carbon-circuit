package service_test

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/carboncircuit/backend/services/sustainability-service/internal/domain"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/service"
)

func figure(t *testing.T, value string) decimal.Decimal {
	t.Helper()

	parsed, err := decimal.NewFromString(value)
	if err != nil {
		t.Fatalf("parse %q: %v", value, err)
	}
	return parsed
}

func TestPriorityFollowsWhatWouldActuallyIssue(t *testing.T) {
	cases := []struct {
		name      string
		requested string
		ceiling   string
		discount  string
		expected  domain.QueuePriority
		dual      bool
	}{
		{
			name:      "modest request at a large facility",
			requested: "3000",
			ceiling:   "5631.632482",
			discount:  "1.00",
			expected:  domain.Normal,
			dual:      false,
		},
		{
			name:      "large request within a larger ceiling",
			requested: "6000",
			ceiling:   "8156.928",
			discount:  "1.00",
			expected:  domain.Critical,
			dual:      true,
		},
		{
			name:      "request over its ceiling, both under the threshold",
			requested: "6125.6",
			ceiling:   "4594.2",
			discount:  "0.75",
			expected:  domain.High,
			dual:      false,
		},
		{
			name:      "request over a ceiling that is itself over the threshold",
			requested: "9000",
			ceiling:   "5631.632482",
			discount:  "1.00",
			expected:  domain.Critical,
			dual:      true,
		},
		{
			name:      "self declared facility within its ceiling",
			requested: "1000",
			ceiling:   "3062.8",
			discount:  "0.50",
			expected:  domain.High,
			dual:      false,
		},
		{
			name:      "exactly at the threshold is not over it",
			requested: "5000",
			ceiling:   "8000",
			discount:  "1.00",
			expected:  domain.Normal,
			dual:      false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			requested := figure(t, testCase.requested)
			ceiling := figure(t, testCase.ceiling)

			priority := service.PriorityFor(requested, ceiling, testCase.discount)
			if priority != testCase.expected {
				t.Errorf("expected %s, got %s", testCase.expected, priority)
			}

			if dual := service.RequiresDualApproval(requested, ceiling); dual != testCase.dual {
				t.Errorf("expected dual approval %v, got %v", testCase.dual, dual)
			}
		})
	}
}

func TestALargeCeilingAloneDoesNotDemandTwoApprovals(t *testing.T) {
	requested := figure(t, "3000")
	ceiling := figure(t, "5631.632482")

	if service.RequiresDualApproval(requested, ceiling) {
		t.Fatal("only 3000 could ever issue here, so one approval is enough under PRD 3.4")
	}
}

func TestTheCeilingStillCapsWhatCanIssue(t *testing.T) {
	requested := figure(t, "9000")
	ceiling := figure(t, "4000")

	if got := service.IssuableAmount(requested, ceiling); !got.Equal(ceiling) {
		t.Fatalf("the ceiling must cap the issuable amount, got %s", got)
	}
}

func TestTheRequestCapsWhatCanIssueWhenItIsLower(t *testing.T) {
	requested := figure(t, "1200")
	ceiling := figure(t, "4000")

	if got := service.IssuableAmount(requested, ceiling); !got.Equal(requested) {
		t.Fatalf("a verifier cannot issue beyond what was asked for, got %s", got)
	}
}
