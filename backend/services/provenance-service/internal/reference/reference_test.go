package reference_test

import (
	"regexp"
	"testing"

	"github.com/carboncircuit/backend/services/provenance-service/internal/reference"
)

var base62 = regexp.MustCompile(`^[0-9A-Za-z]{22}$`)

func TestReferenceIsTwentyTwoBase62Characters(t *testing.T) {
	for attempt := 0; attempt < 200; attempt++ {
		value, err := reference.New()
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		if !base62.MatchString(value) {
			t.Fatalf("reference %q is not 22 base62 characters", value)
		}
	}
}

func TestReferencesDoNotRepeat(t *testing.T) {
	seen := make(map[string]bool, 5000)

	for attempt := 0; attempt < 5000; attempt++ {
		value, err := reference.New()
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		if seen[value] {
			t.Fatalf("reference %q was drawn twice in 5000 draws", value)
		}
		seen[value] = true
	}
}

func TestReferenceIsNotTimeOrdered(t *testing.T) {
	ascending := 0
	previous, err := reference.New()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	const draws = 500
	for attempt := 0; attempt < draws; attempt++ {
		next, err := reference.New()
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		if next > previous {
			ascending++
		}
		previous = next
	}

	if ascending < draws/4 || ascending > draws*3/4 {
		t.Fatalf(
			"references look ordered: %d of %d draws ascended, which would leak creation order",
			ascending, draws,
		)
	}
}
