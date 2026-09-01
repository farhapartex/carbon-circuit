package registry

import "testing"

const registered = "Formosa Precision Semiconductor Co., Ltd."

func TestCosmeticDifferencesStillMatch(t *testing.T) {
	declarations := []string{
		"Formosa Precision Semiconductor Co., Ltd.",
		"Formosa Precision Semiconductor Co Ltd",
		"formosa precision semiconductor co.,ltd",
		"  FORMOSA   PRECISION  SEMICONDUCTOR   CO. LTD.  ",
		"Formosa-Precision-Semiconductor-Co-Ltd",
	}

	for _, declared := range declarations {
		t.Run(declared, func(t *testing.T) {
			if score := NameSimilarity(declared, registered); score < MinimumNameSimilarity {
				t.Fatalf("expected a cosmetic variation to match, scored %.3f", score)
			}
		})
	}
}

func TestTypoIsToleratedButMissingWordsAreNot(t *testing.T) {
	typo := NameSimilarity("Formosa Precision Semiconductors Co., Ltd.", registered)
	if typo < MinimumNameSimilarity {
		t.Fatalf("expected a single-character typo to be tolerated, scored %.3f", typo)
	}

	truncated := NameSimilarity("Formosa Precision Co., Ltd.", registered)
	if truncated >= MinimumNameSimilarity {
		t.Fatalf("expected a dropped significant word to fail, scored %.3f", truncated)
	}
}

func TestUnrelatedNamesFail(t *testing.T) {
	declarations := []string{
		"Meridian Advanced Assembly Co., Ltd.",
		"Formosa",
		"Northwind Integrated Logistics Sdn. Bhd.",
		"",
	}

	for _, declared := range declarations {
		t.Run(declared, func(t *testing.T) {
			if score := NameSimilarity(declared, registered); score >= MinimumNameSimilarity {
				t.Fatalf("expected an unrelated name to fail, scored %.3f", score)
			}
		})
	}
}

func TestSimilarityIsSymmetric(t *testing.T) {
	declared := "Formosa Precision Semiconductors Co Ltd"

	forward := NameSimilarity(declared, registered)
	backward := NameSimilarity(registered, declared)

	if forward != backward {
		t.Fatalf("expected symmetry, got %.3f and %.3f", forward, backward)
	}
}
