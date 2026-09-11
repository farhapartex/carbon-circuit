package creditclass_test

import (
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/services/credit-ledger-service/internal/creditclass"
)

const facility = "01a063a9-a01b-787d-aae9-7da95fd97ac1"

func identifier(t *testing.T, value string) uuid.UUID {
	t.Helper()

	parsed, err := uuid.Parse(value)
	if err != nil {
		t.Fatalf("parse %q: %v", value, err)
	}
	return parsed
}

func TestTheDomainSeparatorMatchesTheRecordedConstant(t *testing.T) {
	domain := creditclass.Domain()

	const recorded = "072975473b061d90de4a27b1c8ed6bcccbeb54b5f8cb03c76cca963aa9287ead"

	if got := hex.EncodeToString(domain[:]); got != recorded {
		t.Fatalf(
			"the domain separator is pinned in the smart contract design as %s but derives as %s; "+
				"changing it re-keys every credit already issued",
			recorded, got)
	}
}

func TestTheSameClassAlwaysDerivesTheSameIdentifier(t *testing.T) {
	first, err := creditclass.TokenID(identifier(t, facility), 2026, creditclass.RenewableEnergy)
	if err != nil {
		t.Fatalf("derive: %v", err)
	}

	second, err := creditclass.TokenID(identifier(t, facility), 2026, creditclass.RenewableEnergy)
	if err != nil {
		t.Fatalf("derive again: %v", err)
	}

	if first != second {
		t.Fatalf("derivation is not deterministic: %s then %s", first, second)
	}
}

func TestEachAttributeChangesTheIdentifier(t *testing.T) {
	base, err := creditclass.TokenID(identifier(t, facility), 2026, creditclass.RenewableEnergy)
	if err != nil {
		t.Fatalf("derive: %v", err)
	}

	cases := map[string]struct {
		facility string
		vintage  int
		activity string
	}{
		"another facility": {"01a063ab-4240-79aa-a0cd-25749825dd43", 2026, creditclass.RenewableEnergy},
		"another vintage":  {facility, 2027, creditclass.RenewableEnergy},
		"another activity": {facility, 2026, creditclass.ResponsibleSourcing},
	}

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			other, err := creditclass.TokenID(
				identifier(t, testCase.facility), testCase.vintage, testCase.activity)
			if err != nil {
				t.Fatalf("derive: %v", err)
			}

			if other == base {
				t.Fatalf("%s produced the same token id, so classes would be pooled", name)
			}
		})
	}
}

func TestTheIdentifierIsAFullWidthUnsignedInteger(t *testing.T) {
	token, err := creditclass.TokenID(identifier(t, facility), 2026, creditclass.RenewableEnergy)
	if err != nil {
		t.Fatalf("derive: %v", err)
	}

	value, ok := new(big.Int).SetString(token, 10)
	if !ok {
		t.Fatalf("token id %q is not a decimal integer", token)
	}
	if value.Sign() <= 0 {
		t.Fatal("a token id derived from a hash should not be zero or negative")
	}
	if value.BitLen() > 256 {
		t.Fatalf("token id needs %d bits, more than uint256 holds", value.BitLen())
	}
}

func TestAnUnknownActivityIsRefused(t *testing.T) {
	_, err := creditclass.TokenID(identifier(t, facility), 2026, "tree_planting")
	if !errors.Is(err, creditclass.ErrActivityUnknown) {
		t.Fatalf("expected ErrActivityUnknown, got %v", err)
	}
}

func TestAVintageOutsideUint16IsRefused(t *testing.T) {
	for _, year := range []int{-1, 65536, 1000000} {
		if _, err := creditclass.TokenID(identifier(t, facility), year, creditclass.RenewableEnergy); err == nil {
			t.Fatalf("vintage %d does not fit a uint16 and must be refused", year)
		}
	}
}

func TestTheDerivationMatchesItsGoldenVectors(t *testing.T) {
	golden := map[string]string{
		creditclass.RenewableEnergy:          "103356341784019533215921733667852901958660570914899948812981618235613999041755",
		creditclass.ReducedEmissionLogistics: "19392662664586271426523246302634003429597238167970888384999504440138579108581",
		creditclass.ResponsibleSourcing:      "104437859222100064792458847054488448823892295187681125589506076859389949949516",
	}

	for activity, expected := range golden {
		t.Run(activity, func(t *testing.T) {
			token, err := creditclass.TokenID(identifier(t, facility), 2026, activity)
			if err != nil {
				t.Fatalf("derive: %v", err)
			}

			if token != expected {
				t.Fatalf(
					"token id for facility %s, vintage 2026, %s must be\n  %s\nbut derived\n  %s\n"+
						"this vector pins the whole preimage: the domain separator, the order of the "+
						"four words, and the alignment of each value within its word. A contract that "+
						"derives anything else cannot mint against balances already issued",
					facility, activity, expected, token)
			}
		})
	}
}

func TestTwoFacilitiesDifferingInOneByteDeriveDifferentIdentifiers(t *testing.T) {
	one := identifier(t, "00000000-0000-0000-0000-000000000001")
	two := identifier(t, "00000000-0000-0000-0000-000000000002")

	first, err := creditclass.TokenID(one, 2026, creditclass.RenewableEnergy)
	if err != nil {
		t.Fatalf("derive: %v", err)
	}

	second, err := creditclass.TokenID(two, 2026, creditclass.RenewableEnergy)
	if err != nil {
		t.Fatalf("derive: %v", err)
	}

	if first == second {
		t.Fatal("two facilities differing only in their last byte must derive different ids")
	}
}
