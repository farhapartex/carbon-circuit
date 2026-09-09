package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	sustainabilityv1 "github.com/carboncircuit/backend/gen/carboncircuit/sustainability/v1"
	"github.com/carboncircuit/backend/internal/httpx"
	"github.com/carboncircuit/backend/services/api-gateway/internal/caller"
)

var decisionOutcomeByName = map[string]sustainabilityv1.DecisionOutcome{
	"approved":                   sustainabilityv1.DecisionOutcome_DECISION_OUTCOME_APPROVED,
	"rejected":                   sustainabilityv1.DecisionOutcome_DECISION_OUTCOME_REJECTED,
	"more_information_requested": sustainabilityv1.DecisionOutcome_DECISION_OUTCOME_MORE_INFORMATION_REQUESTED,
}

var decisionOutcomeName = invertMap(decisionOutcomeByName)

type decideRequest struct {
	Outcome        string `json:"outcome" binding:"required"`
	ApprovedAmount string `json:"approved_amount" binding:"max=32"`
	Reason         string `json:"reason" binding:"max=2000"`
}

type aiReviewResponse struct {
	Assessment       string            `json:"assessment"`
	Confidence       *string           `json:"confidence"`
	ExtractedFigures map[string]string `json:"extracted_figures"`
	Flags            []string          `json:"flags"`
	Narrative        string            `json:"narrative"`
	AssessedBy       string            `json:"assessed_by"`
	AssessedAt       string            `json:"assessed_at"`
}

type decisionResponse struct {
	ID             string  `json:"id"`
	VerifierUserID string  `json:"verifier_user_id"`
	VerifierName   string  `json:"verifier_name"`
	Outcome        string  `json:"outcome"`
	ApprovedAmount *string `json:"approved_amount"`
	Reason         string  `json:"reason"`
	DecidedAt      string  `json:"decided_at"`
}

func toAIReviewResponse(review *sustainabilityv1.AIReview) *aiReviewResponse {
	if review == nil {
		return nil
	}

	return &aiReviewResponse{
		Assessment:       review.GetAssessment(),
		Confidence:       emptyToNil(review.GetConfidence()),
		ExtractedFigures: review.GetExtractedFigures(),
		Flags:            review.GetFlags(),
		Narrative:        review.GetNarrative(),
		AssessedBy:       review.GetAssessedBy(),
		AssessedAt:       review.GetAssessedAt(),
	}
}

func toDecisionResponses(decisions []*sustainabilityv1.ClaimDecision) []decisionResponse {
	recorded := make([]decisionResponse, 0, len(decisions))

	for _, decision := range decisions {
		recorded = append(recorded, decisionResponse{
			ID:             decision.GetId(),
			VerifierUserID: decision.GetVerifierUserId(),
			VerifierName:   decision.GetVerifierName(),
			Outcome:        decisionOutcomeName[decision.GetOutcome()],
			ApprovedAmount: emptyToNil(decision.GetApprovedAmount()),
			Reason:         decision.GetReason(),
			DecidedAt:      decision.GetDecidedAt(),
		})
	}

	return recorded
}

func (h *Handlers) ReviewQueue(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	limit := 0
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
				Field: "limit", Code: "UNSUPPORTED_VALUE",
			})
			return
		}
		limit = parsed
	}

	queued, err := h.Sustainability.ReviewQueue(c.Request.Context(),
		&sustainabilityv1.ReviewQueueRequest{After: c.Query("after"), Limit: int32(limit)})
	if err != nil {
		h.failClaim(c, err)
		return
	}

	claims := make([]claimResponse, 0, len(queued.GetClaims()))
	for _, claim := range queued.GetClaims() {
		claims = append(claims, toClaimResponse(claim))
	}

	httpx.Data(c, http.StatusOK, map[string]any{
		"claims":      claims,
		"next_cursor": emptyToNil(queued.GetNextCursor()),
	})
}

func (h *Handlers) ReviewClaim(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	reviewed, err := h.Sustainability.ReviewClaim(c.Request.Context(), c.Param("claimId"))
	if err != nil {
		h.failClaim(c, err)
		return
	}

	httpx.Data(c, http.StatusOK, map[string]any{
		"claim":     toClaimResponse(reviewed.GetClaim()),
		"ai_review": toAIReviewResponse(reviewed.GetAiReview()),
		"decisions": toDecisionResponses(reviewed.GetDecisions()),
	})
}

func (h *Handlers) DecideClaim(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	idempotencyKey, present := httpx.IdempotencyKeyFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeIdempotencyKeyRequired)
		return
	}

	var body decideRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.Fail(c, httpx.CodeValidation)
		return
	}

	outcome, known := decisionOutcomeByName[body.Outcome]
	if !known {
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
			Field: "outcome", Code: "UNSUPPORTED_VALUE",
		})
		return
	}

	decided, err := h.Sustainability.DecideClaim(c.Request.Context(), idempotencyKey,
		&sustainabilityv1.DecideClaimRequest{
			ClaimId:        c.Param("claimId"),
			Outcome:        outcome,
			ApprovedAmount: body.ApprovedAmount,
			Reason:         body.Reason,
			IdempotencyKey: idempotencyKey,
		})
	if err != nil {
		h.failClaim(c, err)
		return
	}

	httpx.Data(c, http.StatusOK, map[string]any{
		"claim":     toClaimResponse(decided.GetClaim()),
		"decisions": toDecisionResponses(decided.GetDecisions()),
	})
}
