package score

import (
	"sort"
	"time"

	"github.com/carboncircuit/backend/services/provenance-service/internal/domain"
)

const (
	CompletenessPoints = 40
	AnchoringPoints    = 20
	ChainDepthPoints   = 15
	TimelinessPoints   = 15
	FacilityPoints     = 10

	timelyBelow = 24 * time.Hour
	timelyWorst = 7 * 24 * time.Hour
)

type Component struct {
	Label       string `json:"label"`
	Earned      int    `json:"earned"`
	Available   int    `json:"available"`
	Explanation string `json:"explanation"`
}

type Score struct {
	Total      int         `json:"total"`
	Components []Component `json:"components"`
}

type Checkpoint struct {
	Type       domain.CheckpointType
	Anchored   bool
	OccurredAt time.Time
	ReportedAt time.Time
	Superseded bool
}

type FacilityClaim struct {
	VintageYear int
}

type Input struct {
	Category          domain.ProductCategory
	Checkpoints       []Checkpoint
	DeclaredParents   int
	ResolvedParents   int
	FacilityClaims    []FacilityClaim
	FacilityDataKnown bool
	AnchorDataKnown   bool
	Now               time.Time
}

func Compute(input Input) Score {
	live := livingCheckpoints(input.Checkpoints)

	if len(live) == 0 {
		return unstarted(input)
	}

	components := []Component{
		completeness(input.Category, live),
		anchoring(live, input.AnchorDataKnown),
		chainDepth(input.DeclaredParents, input.ResolvedParents),
		timeliness(live),
		facilityRecord(input),
	}

	total := 0
	for _, component := range components {
		total += component.Earned
	}

	return Score{Total: total, Components: components}
}

func livingCheckpoints(checkpoints []Checkpoint) []Checkpoint {
	live := make([]Checkpoint, 0, len(checkpoints))
	for _, checkpoint := range checkpoints {
		if !checkpoint.Superseded {
			live = append(live, checkpoint)
		}
	}
	return live
}

func unstarted(input Input) Score {
	return Score{
		Total: 0,
		Components: []Component{
			{
				Label:       "Checkpoint completeness",
				Available:   CompletenessPoints,
				Explanation: "No checkpoints recorded yet.",
			},
			{
				Label:       "On-chain anchoring",
				Available:   AnchoringPoints,
				Explanation: "Nothing to anchor yet.",
			},
			{
				Label:     "Chain depth resolution",
				Available: ChainDepthPoints,
				Explanation: chainDepthExplanation(
					input.DeclaredParents, input.ResolvedParents,
				),
			},
			{
				Label:       "Reporting timeliness",
				Available:   TimelinessPoints,
				Explanation: "No checkpoints recorded yet.",
			},
			{
				Label:       "Facility sustainability record",
				Available:   FacilityPoints,
				Explanation: facilityExplanation(input, 0),
			},
		},
	}
}

func completeness(
	category domain.ProductCategory,
	checkpoints []Checkpoint,
) Component {
	expected := domain.ExpectedCheckpointSequence[category]

	if len(expected) == 0 {
		return Component{
			Label:       "Checkpoint completeness",
			Earned:      CompletenessPoints,
			Available:   CompletenessPoints,
			Explanation: "No checkpoint sequence is defined for this product category, so completeness cannot count against this batch.",
		}
	}

	recorded := make(map[domain.CheckpointType]bool, len(checkpoints))
	for _, checkpoint := range checkpoints {
		recorded[checkpoint.Type] = true
	}

	present := 0
	for _, wanted := range expected {
		if recorded[wanted] {
			present++
		}
	}

	return Component{
		Label:       "Checkpoint completeness",
		Earned:      present * CompletenessPoints / len(expected),
		Available:   CompletenessPoints,
		Explanation: pluralisedCompleteness(present, len(expected)),
	}
}

func anchoring(checkpoints []Checkpoint, known bool) Component {
	if !known {
		return Component{
			Label:       "On-chain anchoring",
			Available:   AnchoringPoints,
			Explanation: "Anchoring is not being observed yet, so no checkpoint can be credited as anchored.",
		}
	}

	anchored := 0
	for _, checkpoint := range checkpoints {
		if checkpoint.Anchored {
			anchored++
		}
	}

	return Component{
		Label:       "On-chain anchoring",
		Earned:      anchored * AnchoringPoints / len(checkpoints),
		Available:   AnchoringPoints,
		Explanation: pluralisedAnchoring(anchored, len(checkpoints)),
	}
}

func chainDepth(declared, resolved int) Component {
	earned := ChainDepthPoints
	if declared > 0 {
		earned = resolved * ChainDepthPoints / declared
	}

	return Component{
		Label:       "Chain depth resolution",
		Earned:      earned,
		Available:   ChainDepthPoints,
		Explanation: chainDepthExplanation(declared, resolved),
	}
}

func timeliness(checkpoints []Checkpoint) Component {
	median := medianLag(checkpoints)

	earned := 0
	switch {
	case median <= timelyBelow:
		earned = TimelinessPoints
	case median >= timelyWorst:
		earned = 0
	default:
		remaining := timelyWorst - median
		span := timelyWorst - timelyBelow
		earned = int(int64(TimelinessPoints) * int64(remaining) / int64(span))
	}

	return Component{
		Label:       "Reporting timeliness",
		Earned:      earned,
		Available:   TimelinessPoints,
		Explanation: timelinessExplanation(median),
	}
}

func medianLag(checkpoints []Checkpoint) time.Duration {
	lags := make([]time.Duration, 0, len(checkpoints))
	for _, checkpoint := range checkpoints {
		lag := checkpoint.ReportedAt.Sub(checkpoint.OccurredAt)
		if lag < 0 {
			lag = 0
		}
		lags = append(lags, lag)
	}

	sort.Slice(lags, func(a, b int) bool { return lags[a] < lags[b] })

	middle := len(lags) / 2
	if len(lags)%2 == 1 {
		return lags[middle]
	}
	return (lags[middle-1] + lags[middle]) / 2
}

func facilityRecord(input Input) Component {
	if !input.FacilityDataKnown {
		return Component{
			Label:       "Facility sustainability record",
			Available:   FacilityPoints,
			Explanation: facilityExplanation(input, 0),
		}
	}

	earned := 0
	for _, claim := range input.FacilityClaims {
		if input.Now.Year()-claim.VintageYear <= 4 {
			earned = FacilityPoints
			break
		}
		earned = FacilityPoints / 2
	}

	return Component{
		Label:       "Facility sustainability record",
		Earned:      earned,
		Available:   FacilityPoints,
		Explanation: facilityExplanation(input, earned),
	}
}
