package service_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/services/sustainability-service/internal/domain"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/service"
)

const substantiveReason = "The submitted utility statements cover only nine of the twelve claimed months."

func reviewedClaim(requested, ceiling string, dual bool) domain.Claim {
	return domain.Claim{
		RequestedAmount:      requested,
		ComputedCeiling:      ceiling,
		RequiresDualApproval: dual,
		Status:               domain.HumanReview,
	}
}

func approval(amount string) service.Decision {
	return service.Decision{Outcome: domain.DecisionApproved, ApprovedAmount: amount}
}

func TestApprovalCannotExceedTheCeiling(t *testing.T) {
	claim := reviewedClaim("9000.000000", "4594.200000", false)

	err := service.ValidateDecision(approval("5000"), claim)
	if !errors.Is(err, service.ErrAboveCeiling) {
		t.Fatalf("expected ErrAboveCeiling, got %v", err)
	}
}

func TestApprovalCannotExceedWhatWasRequested(t *testing.T) {
	claim := reviewedClaim("400.000000", "5639.998000", false)

	err := service.ValidateDecision(approval("1000"), claim)
	if !errors.Is(err, service.ErrAboveRequested) {
		t.Fatalf("a verifier must not issue more than was asked for, got %v", err)
	}
}

func TestApprovalAtTheCeilingIsAllowed(t *testing.T) {
	claim := reviewedClaim("4594.200000", "4594.200000", false)

	if err := service.ValidateDecision(approval("4594.200000"), claim); err != nil {
		t.Fatalf("approving exactly at the ceiling must be allowed, got %v", err)
	}
}

func TestApprovalBelowTheRequestIsAllowed(t *testing.T) {
	claim := reviewedClaim("400.000000", "5639.998000", false)

	if err := service.ValidateDecision(approval("250"), claim); err != nil {
		t.Fatalf("a verifier may approve less than requested, got %v", err)
	}
}

func TestApprovalMustBePositive(t *testing.T) {
	claim := reviewedClaim("400.000000", "5639.998000", false)

	for _, amount := range []string{"0", "-1", "", "abc"} {
		if err := service.ValidateDecision(approval(amount), claim); !errors.Is(err, service.ErrApprovalNotPositive) {
			t.Fatalf("amount %q should be refused, got %v", amount, err)
		}
	}
}

func TestARejectionNeedsASubstantiveReason(t *testing.T) {
	claim := reviewedClaim("400.000000", "5639.998000", false)

	short := service.Decision{Outcome: domain.DecisionRejected, Reason: "no good"}
	if err := service.ValidateDecision(short, claim); !errors.Is(err, service.ErrReasonTooShort) {
		t.Fatalf("expected ErrReasonTooShort, got %v", err)
	}

	full := service.Decision{Outcome: domain.DecisionRejected, Reason: substantiveReason}
	if err := service.ValidateDecision(full, claim); err != nil {
		t.Fatalf("a substantive reason must be accepted, got %v", err)
	}
}

func TestWhitespaceDoesNotMakeAReasonSubstantive(t *testing.T) {
	claim := reviewedClaim("400.000000", "5639.998000", false)

	padded := service.Decision{
		Outcome: domain.DecisionRejected,
		Reason:  "too thin" + strings.Repeat(" ", 60),
	}

	if err := service.ValidateDecision(padded, claim); !errors.Is(err, service.ErrReasonTooShort) {
		t.Fatalf("padding must not satisfy the reason requirement, got %v", err)
	}
}

func TestAnInformationRequestAlsoNeedsAReason(t *testing.T) {
	claim := reviewedClaim("400.000000", "5639.998000", false)

	thin := service.Decision{Outcome: domain.DecisionMoreInformationRequested, Reason: "send more"}
	if err := service.ValidateDecision(thin, claim); !errors.Is(err, service.ErrReasonTooShort) {
		t.Fatalf("expected ErrReasonTooShort, got %v", err)
	}
}

func decided(outcome domain.DecisionOutcome, amount string, who uuid.UUID) domain.ClaimDecision {
	entry := domain.ClaimDecision{Outcome: outcome, VerifierUserID: who}
	if amount != "" {
		entry.ApprovedAmount = &amount
	}
	return entry
}

func TestASingleApprovalSettlesAnOrdinaryClaim(t *testing.T) {
	claim := reviewedClaim("400.000000", "5639.998000", false)
	decisions := []domain.ClaimDecision{decided(domain.DecisionApproved, "400.000000", uuid.New())}

	status, issued := service.OutcomeOf(claim, decisions)
	if status != domain.Approved {
		t.Fatalf("expected approved, got %s", status)
	}
	if issued == nil || *issued != "400.000000" {
		t.Fatalf("expected 400 issued, got %v", issued)
	}
}

func TestOneApprovalIsNotEnoughWhenTwoAreRequired(t *testing.T) {
	claim := reviewedClaim("6000.000000", "8156.928000", true)
	decisions := []domain.ClaimDecision{decided(domain.DecisionApproved, "6000.000000", uuid.New())}

	status, issued := service.OutcomeOf(claim, decisions)
	if status != domain.HumanReview {
		t.Fatalf("a large claim must wait for a second approval, got %s", status)
	}
	if issued != nil {
		t.Fatal("nothing may be issued on a single approval of a large claim")
	}
}

func TestTwoApprovalsSettleALargeClaimAtTheLowerAmount(t *testing.T) {
	claim := reviewedClaim("6000.000000", "8156.928000", true)
	decisions := []domain.ClaimDecision{
		decided(domain.DecisionApproved, "6000.000000", uuid.New()),
		decided(domain.DecisionApproved, "5200.000000", uuid.New()),
	}

	status, issued := service.OutcomeOf(claim, decisions)
	if status != domain.Approved {
		t.Fatalf("expected approved, got %s", status)
	}
	if issued == nil || *issued != "5200.000000" {
		t.Fatalf("the more cautious approval must win, got %v", issued)
	}
}

func TestARejectionAfterAnApprovalStillRejects(t *testing.T) {
	claim := reviewedClaim("6000.000000", "8156.928000", true)
	decisions := []domain.ClaimDecision{
		decided(domain.DecisionApproved, "6000.000000", uuid.New()),
		decided(domain.DecisionRejected, "", uuid.New()),
	}

	status, issued := service.OutcomeOf(claim, decisions)
	if status != domain.Rejected {
		t.Fatalf("a second verifier refusing must reject the claim, got %s", status)
	}
	if issued != nil {
		t.Fatal("a rejected claim issues nothing")
	}
}

func TestAnInformationRequestPausesTheClaim(t *testing.T) {
	claim := reviewedClaim("400.000000", "5639.998000", false)
	decisions := []domain.ClaimDecision{decided(domain.DecisionMoreInformationRequested, "", uuid.New())}

	status, issued := service.OutcomeOf(claim, decisions)
	if status != domain.MoreInformationRequested {
		t.Fatalf("expected more_information_requested, got %s", status)
	}
	if issued != nil {
		t.Fatal("a paused claim issues nothing")
	}
}

func TestOnlyAVerifierMayReview(t *testing.T) {
	for _, role := range []string{"", "platform_admin", "owner", "admin"} {
		who := service.Verifier{UserID: uuid.New(), PlatformRole: role}
		if err := service.MayReview(who); !errors.Is(err, service.ErrNotAVerifier) {
			t.Fatalf("role %q must not review, got %v", role, err)
		}
	}

	who := service.Verifier{UserID: uuid.New(), PlatformRole: "verifier"}
	if err := service.MayReview(who); err != nil {
		t.Fatalf("a verifier must be allowed to review, got %v", err)
	}
}

func TestAVerifierWithoutAnIdentityMayNotReview(t *testing.T) {
	who := service.Verifier{PlatformRole: "verifier"}

	if err := service.MayReview(who); !errors.Is(err, service.ErrNotAVerifier) {
		t.Fatalf("an unidentified verifier must be refused, got %v", err)
	}
}
