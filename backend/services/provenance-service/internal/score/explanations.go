package score

import (
	"fmt"
	"time"
)

func pluralisedCompleteness(present, expected int) string {
	if present == expected {
		return fmt.Sprintf("All %d expected checkpoint types recorded.", expected)
	}
	return fmt.Sprintf(
		"%d of %d expected checkpoint types recorded.", present, expected,
	)
}

func pluralisedAnchoring(anchored, total int) string {
	if anchored == total {
		return fmt.Sprintf(
			"All %d checkpoints included in a confirmed epoch anchor.", total,
		)
	}
	return fmt.Sprintf(
		"%d of %d checkpoints included in a confirmed epoch anchor.",
		anchored, total,
	)
}

func chainDepthExplanation(declared, resolved int) string {
	if declared == 0 {
		return "This batch declares no parent batches."
	}
	if resolved == declared {
		if declared == 1 {
			return "The declared parent batch resolves to a registered batch."
		}
		return fmt.Sprintf(
			"All %d declared parent batches resolve to registered batches.",
			declared,
		)
	}
	return fmt.Sprintf(
		"%d of %d declared parent batches resolve to a registered batch.",
		resolved, declared,
	)
}

func timelinessExplanation(median time.Duration) string {
	if median < time.Hour {
		return "Median reporting lag under an hour, well under 24 hours."
	}
	if median < 24*time.Hour {
		return fmt.Sprintf(
			"Median reporting lag of %d hours, under 24 hours.",
			int(median.Hours()),
		)
	}
	days := int(median.Hours()) / 24
	if days == 1 {
		return "Median reporting lag of one day."
	}
	return fmt.Sprintf("Median reporting lag of %d days.", days)
}

func facilityExplanation(input Input, earned int) string {
	if !input.FacilityDataKnown {
		return "Sustainability claims are not being tracked yet, so no facility record can be credited."
	}
	switch earned {
	case FacilityPoints:
		return "Originating facility holds an approved claim with a vintage in the last four years."
	case FacilityPoints / 2:
		return "Originating facility holds an approved claim, but its vintage is older than four years."
	default:
		return "Originating facility has no approved sustainability claim."
	}
}
