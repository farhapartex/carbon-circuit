package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sustainabilityv1 "github.com/carboncircuit/backend/gen/carboncircuit/sustainability/v1"
	"github.com/carboncircuit/backend/internal/events"
	"github.com/carboncircuit/backend/internal/httpx"
	"github.com/carboncircuit/backend/services/api-gateway/internal/caller"
)

var activityTypeByName = map[string]sustainabilityv1.ActivityType{
	"renewable_energy":           sustainabilityv1.ActivityType_ACTIVITY_TYPE_RENEWABLE_ENERGY,
	"reduced_emission_logistics": sustainabilityv1.ActivityType_ACTIVITY_TYPE_REDUCED_EMISSION_LOGISTICS,
	"responsible_sourcing":       sustainabilityv1.ActivityType_ACTIVITY_TYPE_RESPONSIBLE_SOURCING,
}

var activityTypeName = invertMap(activityTypeByName)

var claimStatusByName = map[string]sustainabilityv1.ClaimStatus{
	"submitted":                  sustainabilityv1.ClaimStatus_CLAIM_STATUS_SUBMITTED,
	"ai_review":                  sustainabilityv1.ClaimStatus_CLAIM_STATUS_AI_REVIEW,
	"human_review":               sustainabilityv1.ClaimStatus_CLAIM_STATUS_HUMAN_REVIEW,
	"approved":                   sustainabilityv1.ClaimStatus_CLAIM_STATUS_APPROVED,
	"rejected":                   sustainabilityv1.ClaimStatus_CLAIM_STATUS_REJECTED,
	"more_information_requested": sustainabilityv1.ClaimStatus_CLAIM_STATUS_MORE_INFORMATION_REQUESTED,
}

var claimStatusName = invertMap(claimStatusByName)

var priorityName = map[sustainabilityv1.QueuePriority]string{
	sustainabilityv1.QueuePriority_QUEUE_PRIORITY_NORMAL:   "normal",
	sustainabilityv1.QueuePriority_QUEUE_PRIORITY_HIGH:     "high",
	sustainabilityv1.QueuePriority_QUEUE_PRIORITY_CRITICAL: "critical",
}

type submitClaimRequest struct {
	FacilityID          string            `json:"facility_id" binding:"required,uuid"`
	ActivityType        string            `json:"activity_type" binding:"required"`
	VintageYear         int32             `json:"vintage_year" binding:"required"`
	PeriodStart         string            `json:"period_start" binding:"required"`
	PeriodEnd           string            `json:"period_end" binding:"required"`
	DeclaredFigures     map[string]string `json:"declared_figures"`
	RequestedAmount     string            `json:"requested_amount" binding:"required,max=32"`
	EvidenceIDs         []string          `json:"evidence_ids" binding:"required,min=1,max=25,dive,uuid"`
	ExclusivityAttested bool              `json:"exclusivity_attested"`
}

type previewCeilingRequest struct {
	FacilityID   string `json:"facility_id" binding:"required,uuid"`
	ActivityType string `json:"activity_type" binding:"required"`
	VintageYear  int32  `json:"vintage_year" binding:"required"`
	PeriodStart  string `json:"period_start" binding:"required"`
	PeriodEnd    string `json:"period_end" binding:"required"`
}

type claimEvidenceResponse struct {
	EvidenceID  string `json:"evidence_id"`
	FileName    string `json:"file_name"`
	MediaType   string `json:"media_type"`
	ContentHash string `json:"content_hash"`
	PageCount   *int32 `json:"page_count"`
	ByteSize    int64  `json:"byte_size"`
}

type claimResponse struct {
	ID                    string                  `json:"id"`
	FacilityID            string                  `json:"facility_id"`
	FacilityName          string                  `json:"facility_name"`
	ActivityType          string                  `json:"activity_type"`
	VintageYear           int32                   `json:"vintage_year"`
	PeriodStart           string                  `json:"period_start"`
	PeriodEnd             string                  `json:"period_end"`
	DeclaredFigures       map[string]string       `json:"declared_figures"`
	RequestedAmount       string                  `json:"requested_amount"`
	ComputedCeiling       string                  `json:"computed_ceiling"`
	CapacityBasis         string                  `json:"capacity_basis"`
	CapacitySource        string                  `json:"capacity_source"`
	DiscountFactor        string                  `json:"discount_factor"`
	ReferenceFactorValue  string                  `json:"reference_factor_value"`
	ReferenceFactorID     string                  `json:"reference_factor_id"`
	ReferenceLookupKey    string                  `json:"reference_lookup_key"`
	Status                string                  `json:"status"`
	Priority              string                  `json:"priority"`
	RequiresDualApproval  bool                    `json:"requires_dual_approval"`
	ExclusivityAttestedAt string                  `json:"exclusivity_attested_at"`
	IssuedAmount          *string                 `json:"issued_amount"`
	CreatedAt             string                  `json:"created_at"`
	Evidence              []claimEvidenceResponse `json:"evidence"`
}

func toClaimResponse(claim *sustainabilityv1.Claim) claimResponse {
	evidence := make([]claimEvidenceResponse, 0, len(claim.GetEvidence()))
	for _, attachment := range claim.GetEvidence() {
		var pages *int32
		if attachment.GetPageCount() > 0 {
			counted := attachment.GetPageCount()
			pages = &counted
		}
		evidence = append(evidence, claimEvidenceResponse{
			EvidenceID:  attachment.GetEvidenceId(),
			FileName:    attachment.GetFileName(),
			MediaType:   attachment.GetMediaType(),
			ContentHash: attachment.GetContentHash(),
			PageCount:   pages,
			ByteSize:    attachment.GetByteSize(),
		})
	}

	return claimResponse{
		ID:                    claim.GetId(),
		FacilityID:            claim.GetFacilityId(),
		FacilityName:          claim.GetFacilityName(),
		ActivityType:          activityTypeName[claim.GetActivityType()],
		VintageYear:           claim.GetVintageYear(),
		PeriodStart:           claim.GetPeriodStart(),
		PeriodEnd:             claim.GetPeriodEnd(),
		DeclaredFigures:       claim.GetDeclaredFigures(),
		RequestedAmount:       claim.GetRequestedAmount(),
		ComputedCeiling:       claim.GetComputedCeiling(),
		CapacityBasis:         claim.GetCapacityBasis(),
		CapacitySource:        claim.GetCapacitySource(),
		DiscountFactor:        claim.GetDiscountFactor(),
		ReferenceFactorValue:  claim.GetReferenceFactorValue(),
		ReferenceFactorID:     claim.GetReferenceFactorId(),
		ReferenceLookupKey:    claim.GetReferenceLookupKey(),
		Status:                claimStatusName[claim.GetStatus()],
		Priority:              priorityName[claim.GetPriority()],
		RequiresDualApproval:  claim.GetRequiresDualApproval(),
		ExclusivityAttestedAt: claim.GetExclusivityAttestedAt(),
		IssuedAmount:          emptyToNil(claim.GetIssuedAmount()),
		CreatedAt:             claim.GetCreatedAt(),
		Evidence:              evidence,
	}
}

func (h *Handlers) SubmitClaim(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	idempotencyKey, present := httpx.IdempotencyKeyFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeIdempotencyKeyRequired)
		return
	}

	var body submitClaimRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.Fail(c, httpx.CodeValidation)
		return
	}

	activity, known := activityTypeByName[body.ActivityType]
	if !known {
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
			Field: "activity_type", Code: "UNSUPPORTED_VALUE",
		})
		return
	}

	if !body.ExclusivityAttested {
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
			Field: "exclusivity_attested", Code: "REQUIRED",
		})
		return
	}

	submitted, err := h.Sustainability.SubmitClaim(c.Request.Context(), idempotencyKey,
		&sustainabilityv1.SubmitClaimRequest{
			FacilityId:          body.FacilityID,
			ActivityType:        activity,
			VintageYear:         body.VintageYear,
			PeriodStart:         body.PeriodStart,
			PeriodEnd:           body.PeriodEnd,
			DeclaredFigures:     body.DeclaredFigures,
			RequestedAmount:     body.RequestedAmount,
			EvidenceIds:         body.EvidenceIDs,
			ExclusivityAttested: body.ExclusivityAttested,
			IdempotencyKey:      idempotencyKey,
		})
	if err != nil {
		h.failClaim(c, err)
		return
	}

	statusCode := http.StatusCreated
	if submitted.GetAlreadySubmitted() {
		statusCode = http.StatusOK
	}

	httpx.Data(c, statusCode, toClaimResponse(submitted.GetClaim()))
}

func (h *Handlers) ListClaims(c *gin.Context) {
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

	listed, err := h.Sustainability.ListClaims(c.Request.Context(),
		&sustainabilityv1.ListClaimsRequest{
			Status: claimStatusByName[c.Query("status")],
			After:  c.Query("after"),
			Limit:  int32(limit),
		})
	if err != nil {
		h.failClaim(c, err)
		return
	}

	claims := make([]claimResponse, 0, len(listed.GetClaims()))
	for _, claim := range listed.GetClaims() {
		claims = append(claims, toClaimResponse(claim))
	}

	httpx.Data(c, http.StatusOK, map[string]any{
		"claims":      claims,
		"next_cursor": emptyToNil(listed.GetNextCursor()),
	})
}

func (h *Handlers) GetClaim(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	found, err := h.Sustainability.GetClaim(c.Request.Context(), c.Param("claimId"))
	if err != nil {
		h.failClaim(c, err)
		return
	}

	httpx.Data(c, http.StatusOK, toClaimResponse(found.GetClaim()))
}

func (h *Handlers) PreviewClaimCeiling(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	var body previewCeilingRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.Fail(c, httpx.CodeValidation)
		return
	}

	activity, known := activityTypeByName[body.ActivityType]
	if !known {
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
			Field: "activity_type", Code: "UNSUPPORTED_VALUE",
		})
		return
	}

	preview, err := h.Sustainability.PreviewCeiling(c.Request.Context(),
		&sustainabilityv1.PreviewCeilingRequest{
			FacilityId:   body.FacilityID,
			ActivityType: activity,
			VintageYear:  body.VintageYear,
			PeriodStart:  body.PeriodStart,
			PeriodEnd:    body.PeriodEnd,
		})
	if err != nil {
		h.failClaim(c, err)
		return
	}

	httpx.Data(c, http.StatusOK, map[string]any{
		"ceiling":         preview.GetCeiling(),
		"capacity_basis":  preview.GetCapacityBasis(),
		"capacity_source": preview.GetCapacitySource(),
		"discount_factor": preview.GetDiscountFactor(),
		"reference_value": preview.GetReferenceValue(),
		"grid_region":     preview.GetGridRegion(),
		"period_days":     preview.GetPeriodDays(),
		"vintage_days":    preview.GetVintageDays(),
	})
}

func claimRefusalReason(err error) string {
	for _, detail := range status.Convert(err).Details() {
		info, matches := detail.(*errdetails.ErrorInfo)
		if matches && info.GetDomain() == events.ClaimRefusalDomain {
			return info.GetReason()
		}
	}
	return "REFUSED"
}

func (h *Handlers) failClaim(c *gin.Context, err error) {
	reported := status.Convert(err)

	switch reported.Code() {
	case codes.NotFound:
		httpx.Fail(c, httpx.CodeResourceNotFound)
	case codes.PermissionDenied:
		httpx.Fail(c, httpx.CodeOrganizationReadOnly)
	case codes.FailedPrecondition:
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
			Field: "claim", Code: claimRefusalReason(err),
		})
	case codes.InvalidArgument:
		httpx.Fail(c, httpx.CodeValidation)
	case codes.AlreadyExists:
		httpx.Fail(c, httpx.CodeIdempotencyKeyReused)
	case codes.Aborted:
		httpx.Fail(c, httpx.CodeRequestInProgress)
	default:
		h.Logger.Error("claim call failed", errorAttributes(c, err)...)
		httpx.Fail(c, httpx.CodeDependencyUnavailable)
	}
}
