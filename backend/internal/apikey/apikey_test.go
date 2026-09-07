package apikey_test

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/carboncircuit/backend/internal/apikey"
)

const pepper = "a-development-pepper-value"

var presentedShape = regexp.MustCompile(`^cc_live_[0-9a-z]{8}_[A-Za-z0-9_-]{43}$`)

func hasher(t *testing.T) *apikey.Hasher {
	t.Helper()

	built, err := apikey.NewHasher(pepper)
	if err != nil {
		t.Fatalf("build hasher: %v", err)
	}
	return built
}

func TestHasherRefusesAnEmptyPepper(t *testing.T) {
	for _, blank := range []string{"", "   "} {
		if _, err := apikey.NewHasher(blank); !errors.Is(err, apikey.ErrPepperUnset) {
			t.Fatalf("expected ErrPepperUnset for %q, got %v", blank, err)
		}
	}
}

func TestIssuedKeyHasTheExpectedShape(t *testing.T) {
	issued, err := hasher(t).Issue()
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	if !presentedShape.MatchString(issued.Presented()) {
		t.Fatalf("unexpected key shape: %q", issued.Presented())
	}
	if len(issued.Prefix) != apikey.PrefixLength {
		t.Fatalf("expected an %d character prefix, got %q", apikey.PrefixLength, issued.Prefix)
	}
}

func TestTheStoredHashNeverContainsTheSecret(t *testing.T) {
	issued, err := hasher(t).Issue()
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	if strings.Contains(string(issued.HMAC), issued.Secret) {
		t.Fatal("the stored hash must not contain the secret")
	}
	if len(issued.HMAC) != 32 {
		t.Fatalf("expected a 32 byte hmac, got %d", len(issued.HMAC))
	}
}

func TestKeysDoNotRepeat(t *testing.T) {
	built := hasher(t)
	prefixes := make(map[string]bool, 2000)
	secrets := make(map[string]bool, 2000)

	for attempt := 0; attempt < 2000; attempt++ {
		issued, err := built.Issue()
		if err != nil {
			t.Fatalf("issue: %v", err)
		}
		if prefixes[issued.Prefix] {
			t.Fatalf("prefix %q was drawn twice", issued.Prefix)
		}
		if secrets[issued.Secret] {
			t.Fatal("a secret was drawn twice")
		}
		prefixes[issued.Prefix] = true
		secrets[issued.Secret] = true
	}
}

func TestMatchesAcceptsTheIssuedSecretAndRejectsOthers(t *testing.T) {
	built := hasher(t)

	issued, err := built.Issue()
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	if !built.Matches(issued.Secret, issued.HMAC) {
		t.Fatal("the issued secret must verify against its own hash")
	}

	other, err := built.Issue()
	if err != nil {
		t.Fatalf("issue second: %v", err)
	}

	if built.Matches(other.Secret, issued.HMAC) {
		t.Fatal("a different secret must not verify")
	}
	if built.Matches(issued.Secret+"x", issued.HMAC) {
		t.Fatal("a mutated secret must not verify")
	}
}

func TestADifferentPepperNeverVerifies(t *testing.T) {
	issued, err := hasher(t).Issue()
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	stolen, err := apikey.NewHasher("a-different-pepper")
	if err != nil {
		t.Fatalf("build second hasher: %v", err)
	}

	if stolen.Matches(issued.Secret, issued.HMAC) {
		t.Fatal("a hash from one pepper must not verify under another")
	}
}

func TestParseRoundTripsAnIssuedKey(t *testing.T) {
	issued, err := hasher(t).Issue()
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	parsed, err := apikey.Parse(issued.Presented())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if parsed.Prefix != issued.Prefix || parsed.Secret != issued.Secret {
		t.Fatalf("round trip lost data: %+v", parsed)
	}
}

func TestParseRejectsMalformedKeys(t *testing.T) {
	cases := map[string]string{
		"empty":          "",
		"no scheme":      "abcdefgh_secret",
		"wrong scheme":   "cc_test_abcdefgh_secret",
		"short prefix":   "cc_live_abc_secret",
		"missing secret": "cc_live_abcdefgh_",
		"only scheme":    "cc_live",
	}

	for name, candidate := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := apikey.Parse(candidate); !errors.Is(err, apikey.ErrMalformed) {
				t.Fatalf("expected ErrMalformed for %q, got %v", candidate, err)
			}
		})
	}
}

func TestParseKeepsUnderscoresInsideTheSecret(t *testing.T) {
	parsed, err := apikey.Parse("cc_live_abcdefgh_has_underscores_inside")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if parsed.Secret != "has_underscores_inside" {
		t.Fatalf("the secret was truncated at an underscore: %q", parsed.Secret)
	}
}
