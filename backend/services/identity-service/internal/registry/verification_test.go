package registry

import (
	"testing"

	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
)

func TestRegisteredNameIsDisclosedOnlyWhenActionable(t *testing.T) {
	dissolved := domain.RejectionEntityDissolved
	sanctioned := domain.RejectionSanctionsFlag
	mismatch := domain.RejectionNameMismatch

	cases := map[string]struct {
		outcome  Outcome
		expected string
	}{
		"verified discloses the matched name": {
			outcome:  Outcome{Status: domain.VerificationVerified, RegisteredName: registered},
			expected: registered,
		},
		"name mismatch discloses so the user can correct it": {
			outcome: Outcome{
				Status: domain.VerificationRejected, Rejection: &mismatch, RegisteredName: registered,
			},
			expected: registered,
		},
		"dissolved withholds the record": {
			outcome: Outcome{
				Status: domain.VerificationRejected, Rejection: &dissolved, RegisteredName: registered,
			},
			expected: "",
		},
		"sanctioned withholds the record": {
			outcome: Outcome{
				Status: domain.VerificationRejected, Rejection: &sanctioned, RegisteredName: registered,
			},
			expected: "",
		},
		"unverified has no record to disclose": {
			outcome:  Outcome{Status: domain.VerificationUnverified},
			expected: "",
		},
	}

	for name, expectation := range cases {
		t.Run(name, func(t *testing.T) {
			if disclosed := expectation.outcome.DisclosableName(); disclosed != expectation.expected {
				t.Fatalf("expected %q, got %q", expectation.expected, disclosed)
			}
		})
	}
}

func TestAssessCarriesTheRegisteredName(t *testing.T) {
	record := domain.BusinessRegistryRecord{
		LegalName:    registered,
		EntityStatus: domain.RegistryActive,
	}

	outcome := Assess(Declaration{LegalName: registered}, record)

	if outcome.RegisteredName != registered {
		t.Fatalf("expected the registered name to be carried, got %q", outcome.RegisteredName)
	}
	if outcome.Status != domain.VerificationVerified {
		t.Fatalf("expected verified, got %q", outcome.Status)
	}
}
