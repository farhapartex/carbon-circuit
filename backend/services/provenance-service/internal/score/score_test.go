package score_test

import (
	"testing"
	"time"

	"github.com/carboncircuit/backend/services/provenance-service/internal/domain"
	"github.com/carboncircuit/backend/services/provenance-service/internal/score"
)

var reference = time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)

func checkpointAt(
	kind domain.CheckpointType,
	lag time.Duration,
	anchored bool,
) score.Checkpoint {
	occurred := reference.Add(-30 * 24 * time.Hour)
	return score.Checkpoint{
		Type:       kind,
		Anchored:   anchored,
		OccurredAt: occurred,
		ReportedAt: occurred.Add(lag),
	}
}

func componentByLabel(result score.Score, label string) score.Component {
	for _, component := range result.Components {
		if component.Label == label {
			return component
		}
	}
	return score.Component{}
}

func TestBatchWithNoCheckpointsScoresZero(t *testing.T) {
	result := score.Compute(score.Input{
		Category:          domain.CategoryElectronics,
		FacilityDataKnown: true,
		FacilityClaims:    []score.FacilityClaim{{VintageYear: 2026}},
		Now:               reference,
	})

	if result.Total != 0 {
		t.Fatalf("PRD 2.4 requires a batch with no checkpoints to score 0, got %d", result.Total)
	}
	for _, component := range result.Components {
		if component.Earned != 0 {
			t.Fatalf("component %q earned %d on an unstarted batch", component.Label, component.Earned)
		}
	}
}

func TestCompletenessScalesWithDistinctExpectedTypes(t *testing.T) {
	three := []score.Checkpoint{
		checkpointAt(domain.ProductionComplete, time.Hour, false),
		checkpointAt(domain.DepartedOrigin, time.Hour, false),
		checkpointAt(domain.CustomsExport, time.Hour, false),
	}

	result := score.Compute(score.Input{
		Category:    domain.CategoryElectronics,
		Checkpoints: three,
		Now:         reference,
	})

	completeness := componentByLabel(result, "Checkpoint completeness")
	if completeness.Earned != 24 {
		t.Fatalf("expected 3 of 5 types to earn 24 of 40, got %d", completeness.Earned)
	}
}

func TestRepeatedTypesDoNotInflateCompleteness(t *testing.T) {
	repeated := []score.Checkpoint{
		checkpointAt(domain.DepartedOrigin, time.Hour, false),
		checkpointAt(domain.DepartedOrigin, time.Hour, false),
		checkpointAt(domain.DepartedOrigin, time.Hour, false),
		checkpointAt(domain.DepartedOrigin, time.Hour, false),
		checkpointAt(domain.DepartedOrigin, time.Hour, false),
	}

	result := score.Compute(score.Input{
		Category:    domain.CategoryElectronics,
		Checkpoints: repeated,
		Now:         reference,
	})

	completeness := componentByLabel(result, "Checkpoint completeness")
	if completeness.Earned != 8 {
		t.Fatalf("five copies of one type is one distinct type, expected 8 of 40, got %d", completeness.Earned)
	}
}

func TestCategoryWithNoDefinedSequenceIsNotPenalised(t *testing.T) {
	result := score.Compute(score.Input{
		Category: domain.CategoryTextiles,
		Checkpoints: []score.Checkpoint{
			checkpointAt(domain.ProductionComplete, time.Hour, false),
		},
		Now: reference,
	})

	completeness := componentByLabel(result, "Checkpoint completeness")
	if completeness.Earned != score.CompletenessPoints {
		t.Fatalf("an undefined sequence must not penalise the batch, got %d", completeness.Earned)
	}
}

func TestAnchoringScalesAndIsZeroWhenUnobserved(t *testing.T) {
	checkpoints := []score.Checkpoint{
		checkpointAt(domain.ProductionComplete, time.Hour, true),
		checkpointAt(domain.DepartedOrigin, time.Hour, true),
		checkpointAt(domain.CustomsExport, time.Hour, false),
	}

	observed := score.Compute(score.Input{
		Category:        domain.CategoryElectronics,
		Checkpoints:     checkpoints,
		AnchorDataKnown: true,
		Now:             reference,
	})
	if got := componentByLabel(observed, "On-chain anchoring").Earned; got != 13 {
		t.Fatalf("expected 2 of 3 anchored to earn 13 of 20, got %d", got)
	}

	unobserved := score.Compute(score.Input{
		Category:    domain.CategoryElectronics,
		Checkpoints: checkpoints,
		Now:         reference,
	})
	if got := componentByLabel(unobserved, "On-chain anchoring").Earned; got != 0 {
		t.Fatalf("expected no anchoring credit while unobserved, got %d", got)
	}
}

func TestChainDepthIsFullWithNoParentsAndScalesOtherwise(t *testing.T) {
	one := []score.Checkpoint{checkpointAt(domain.ProductionComplete, time.Hour, false)}

	none := score.Compute(score.Input{
		Category: domain.CategoryElectronics, Checkpoints: one, Now: reference,
	})
	if got := componentByLabel(none, "Chain depth resolution").Earned; got != 15 {
		t.Fatalf("declaring no parents earns the full 15, got %d", got)
	}

	partial := score.Compute(score.Input{
		Category:        domain.CategoryElectronics,
		Checkpoints:     one,
		DeclaredParents: 3,
		ResolvedParents: 1,
		Now:             reference,
	})
	if got := componentByLabel(partial, "Chain depth resolution").Earned; got != 5 {
		t.Fatalf("expected 1 of 3 resolved to earn 5 of 15, got %d", got)
	}
}

func TestTimelinessBoundaries(t *testing.T) {
	cases := []struct {
		name   string
		lag    time.Duration
		earned int
	}{
		{"under a day", 3 * time.Hour, 15},
		{"exactly a day", 24 * time.Hour, 15},
		{"seven days", 7 * 24 * time.Hour, 0},
		{"beyond seven days", 30 * 24 * time.Hour, 0},
		{"midpoint", 96 * time.Hour, 7},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			result := score.Compute(score.Input{
				Category: domain.CategoryElectronics,
				Checkpoints: []score.Checkpoint{
					checkpointAt(domain.ProductionComplete, test.lag, false),
				},
				Now: reference,
			})

			if got := componentByLabel(result, "Reporting timeliness").Earned; got != test.earned {
				t.Fatalf("lag %s expected %d, got %d", test.lag, test.earned, got)
			}
		})
	}
}

func TestSupersededCheckpointsAreExcluded(t *testing.T) {
	checkpoints := []score.Checkpoint{
		checkpointAt(domain.ProductionComplete, time.Hour, false),
		{
			Type:       domain.DepartedOrigin,
			OccurredAt: reference.Add(-time.Hour),
			ReportedAt: reference,
			Superseded: true,
		},
	}

	result := score.Compute(score.Input{
		Category:    domain.CategoryElectronics,
		Checkpoints: checkpoints,
		Now:         reference,
	})

	completeness := componentByLabel(result, "Checkpoint completeness")
	if completeness.Earned != 8 {
		t.Fatalf("a superseded checkpoint must not count toward completeness, got %d", completeness.Earned)
	}
}

func TestFacilityRecordRespectsVintageAge(t *testing.T) {
	one := []score.Checkpoint{checkpointAt(domain.ProductionComplete, time.Hour, false)}

	recent := score.Compute(score.Input{
		Category:          domain.CategoryElectronics,
		Checkpoints:       one,
		FacilityDataKnown: true,
		FacilityClaims:    []score.FacilityClaim{{VintageYear: 2024}},
		Now:               reference,
	})
	if got := componentByLabel(recent, "Facility sustainability record").Earned; got != 10 {
		t.Fatalf("a claim within four years earns 10, got %d", got)
	}

	stale := score.Compute(score.Input{
		Category:          domain.CategoryElectronics,
		Checkpoints:       one,
		FacilityDataKnown: true,
		FacilityClaims:    []score.FacilityClaim{{VintageYear: 2015}},
		Now:               reference,
	})
	if got := componentByLabel(stale, "Facility sustainability record").Earned; got != 5 {
		t.Fatalf("a claim older than four years earns 5, got %d", got)
	}
}

func TestPerfectBatchTodayCapsAtSeventy(t *testing.T) {
	every := []score.Checkpoint{
		checkpointAt(domain.ProductionComplete, time.Hour, false),
		checkpointAt(domain.DepartedOrigin, time.Hour, false),
		checkpointAt(domain.CustomsExport, time.Hour, false),
		checkpointAt(domain.CustomsImport, time.Hour, false),
		checkpointAt(domain.ArrivedDestination, time.Hour, false),
	}

	result := score.Compute(score.Input{
		Category:    domain.CategoryElectronics,
		Checkpoints: every,
		Now:         reference,
	})

	if result.Total != 70 {
		t.Fatalf("without anchoring or claim data a flawless batch tops out at 70, got %d", result.Total)
	}
}

func TestFullyCreditedBatchReachesOneHundred(t *testing.T) {
	every := []score.Checkpoint{
		checkpointAt(domain.ProductionComplete, time.Hour, true),
		checkpointAt(domain.DepartedOrigin, time.Hour, true),
		checkpointAt(domain.CustomsExport, time.Hour, true),
		checkpointAt(domain.CustomsImport, time.Hour, true),
		checkpointAt(domain.ArrivedDestination, time.Hour, true),
	}

	result := score.Compute(score.Input{
		Category:          domain.CategoryElectronics,
		Checkpoints:       every,
		AnchorDataKnown:   true,
		FacilityDataKnown: true,
		FacilityClaims:    []score.FacilityClaim{{VintageYear: 2026}},
		Now:               reference,
	})

	if result.Total != 100 {
		t.Fatalf("expected a fully credited batch to reach 100, got %d", result.Total)
	}
}
