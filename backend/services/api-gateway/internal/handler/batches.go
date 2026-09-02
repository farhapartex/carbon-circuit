package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	provenancev1 "github.com/carboncircuit/backend/gen/carboncircuit/provenance/v1"
	"github.com/carboncircuit/backend/internal/httpx"
	"github.com/carboncircuit/backend/services/api-gateway/internal/caller"
)

var batchCategoryByName = map[string]provenancev1.ProductCategory{
	"electronics": provenancev1.ProductCategory_PRODUCT_CATEGORY_ELECTRONICS,
	"agriculture": provenancev1.ProductCategory_PRODUCT_CATEGORY_AGRICULTURE,
	"pharma":      provenancev1.ProductCategory_PRODUCT_CATEGORY_PHARMA,
	"textiles":    provenancev1.ProductCategory_PRODUCT_CATEGORY_TEXTILES,
}

var batchCategoryName = invertMap(batchCategoryByName)

var checkpointTypeByName = map[string]provenancev1.CheckpointType{
	"production_complete": provenancev1.CheckpointType_CHECKPOINT_TYPE_PRODUCTION_COMPLETE,
	"departed_origin":     provenancev1.CheckpointType_CHECKPOINT_TYPE_DEPARTED_ORIGIN,
	"customs_export":      provenancev1.CheckpointType_CHECKPOINT_TYPE_CUSTOMS_EXPORT,
	"customs_import":      provenancev1.CheckpointType_CHECKPOINT_TYPE_CUSTOMS_IMPORT,
	"arrived_destination": provenancev1.CheckpointType_CHECKPOINT_TYPE_ARRIVED_DESTINATION,
}

var checkpointTypeName = invertMap(checkpointTypeByName)

var anchorStatusName = map[provenancev1.AnchorStatus]string{
	provenancev1.AnchorStatus_ANCHOR_STATUS_UNANCHORED:  "unanchored",
	provenancev1.AnchorStatus_ANCHOR_STATUS_PROVISIONAL: "provisional",
	provenancev1.AnchorStatus_ANCHOR_STATUS_CONFIRMED:   "confirmed",
}

type scoreComponentResponse struct {
	Label       string `json:"label"`
	Earned      int32  `json:"earned"`
	Available   int32  `json:"available"`
	Explanation string `json:"explanation"`
}

type provenanceScoreResponse struct {
	Total      int32                    `json:"total"`
	Components []scoreComponentResponse `json:"components"`
}

type batchResponse struct {
	ID                      string                  `json:"id"`
	OriginatingFacilityID   string                  `json:"originating_facility_id"`
	OriginatingFacilityName string                  `json:"originating_facility_name"`
	PublicReference         string                  `json:"public_reference"`
	ProductCategory         string                  `json:"product_category"`
	ComponentType           string                  `json:"component_type"`
	LotNumber               *string                 `json:"lot_number"`
	Quantity                string                  `json:"quantity"`
	Unit                    string                  `json:"unit"`
	ProducedAt              string                  `json:"produced_at"`
	ExternalID              *string                 `json:"external_id"`
	CheckpointCount         int32                   `json:"checkpoint_count"`
	ProvenanceScore         provenanceScoreResponse `json:"provenance_score"`
	CreatedAt               string                  `json:"created_at"`
}

type parentBatchResponse struct {
	DeclaredReference       string  `json:"declared_reference"`
	Resolved                bool    `json:"resolved"`
	ID                      *string `json:"id"`
	ComponentType           *string `json:"component_type"`
	ProductCategory         *string `json:"product_category"`
	OriginatingFacilityName *string `json:"originating_facility_name"`
}

type checkpointResponse struct {
	ID                         string  `json:"id"`
	BatchID                    string  `json:"batch_id"`
	Type                       string  `json:"type"`
	LocationLabel              string  `json:"location_label"`
	CountryCode                string  `json:"country_code"`
	Latitude                   *string `json:"latitude"`
	Longitude                  *string `json:"longitude"`
	ShippingMethod             *string `json:"shipping_method"`
	OccurredAt                 string  `json:"occurred_at"`
	ReportedAt                 string  `json:"reported_at"`
	ReportedByOrganizationName string  `json:"reported_by_organization_name"`
	AnchorStatus               string  `json:"anchor_status"`
	AnchorEpoch                *int32  `json:"anchor_epoch"`
	AnchorTransactionHash      *string `json:"anchor_transaction_hash"`
	InclusionProofAvailable    bool    `json:"inclusion_proof_available"`
	SupersedesCheckpointID     *string `json:"supersedes_checkpoint_id"`
	SupersededByCheckpointID   *string `json:"superseded_by_checkpoint_id"`
	CorrectionReason           *string `json:"correction_reason"`
}

type createBatchRequest struct {
	OriginatingFacilityID string   `json:"originating_facility_id" binding:"required,uuid"`
	ProductCategory       string   `json:"product_category" binding:"required"`
	ComponentType         string   `json:"component_type" binding:"required,max=160"`
	LotNumber             string   `json:"lot_number" binding:"max=64"`
	Quantity              string   `json:"quantity" binding:"required,max=32"`
	Unit                  string   `json:"unit" binding:"required,max=32"`
	ProducedAt            string   `json:"produced_at" binding:"required"`
	ExternalID            string   `json:"external_id" binding:"max=128"`
	ParentReferences      []string `json:"parent_references" binding:"max=25,dive,len=22"`
}

type logCheckpointRequest struct {
	Type                 string `json:"type" binding:"required"`
	LocationLabel        string `json:"location_label" binding:"required,max=120"`
	CountryCode          string `json:"country_code" binding:"required,len=2"`
	Latitude             string `json:"latitude" binding:"max=16"`
	Longitude            string `json:"longitude" binding:"max=16"`
	ShippingMethod       string `json:"shipping_method" binding:"max=32"`
	OccurredAt           string `json:"occurred_at" binding:"required"`
	CorrectsCheckpointID string `json:"corrects_checkpoint_id" binding:"omitempty,uuid"`
	CorrectionReason     string `json:"correction_reason" binding:"max=500"`
	ExternalID           string `json:"external_id" binding:"max=128"`
}

func toScoreResponse(
	scored *provenancev1.ProvenanceScore,
) provenanceScoreResponse {
	components := make([]scoreComponentResponse, 0, len(scored.GetComponents()))
	for _, component := range scored.GetComponents() {
		components = append(components, scoreComponentResponse{
			Label:       component.GetLabel(),
			Earned:      component.GetEarned(),
			Available:   component.GetAvailable(),
			Explanation: component.GetExplanation(),
		})
	}

	return provenanceScoreResponse{
		Total:      scored.GetTotal(),
		Components: components,
	}
}

func toBatchResponse(batch *provenancev1.Batch) batchResponse {
	return batchResponse{
		ID:                      batch.GetId(),
		OriginatingFacilityID:   batch.GetOriginatingFacilityId(),
		OriginatingFacilityName: batch.GetOriginatingFacilityName(),
		PublicReference:         batch.GetPublicReference(),
		ProductCategory:         batchCategoryName[batch.GetProductCategory()],
		ComponentType:           batch.GetComponentType(),
		LotNumber:               emptyToNil(batch.GetLotNumber()),
		Quantity:                batch.GetQuantity(),
		Unit:                    batch.GetUnit(),
		ProducedAt:              batch.GetProducedAt(),
		ExternalID:              emptyToNil(batch.GetExternalId()),
		CheckpointCount:         batch.GetCheckpointCount(),
		ProvenanceScore:         toScoreResponse(batch.GetProvenanceScore()),
		CreatedAt:               batch.GetCreatedAt(),
	}
}

func toParentResponses(
	parents []*provenancev1.ParentBatch,
) []parentBatchResponse {
	mapped := make([]parentBatchResponse, 0, len(parents))

	for _, parent := range parents {
		entry := parentBatchResponse{
			DeclaredReference: parent.GetDeclaredReference(),
			Resolved:          parent.GetResolved(),
			ID:                emptyToNil(parent.GetId()),
			ComponentType:     emptyToNil(parent.GetComponentType()),
			OriginatingFacilityName: emptyToNil(
				parent.GetOriginatingFacilityName(),
			),
		}

		if name := batchCategoryName[parent.GetProductCategory()]; name != "" {
			entry.ProductCategory = &name
		}

		mapped = append(mapped, entry)
	}

	return mapped
}

func toCheckpointResponse(
	checkpoint *provenancev1.Checkpoint,
) checkpointResponse {
	mapped := checkpointResponse{
		ID:                         checkpoint.GetId(),
		BatchID:                    checkpoint.GetBatchId(),
		Type:                       checkpointTypeName[checkpoint.GetType()],
		LocationLabel:              checkpoint.GetLocationLabel(),
		CountryCode:                checkpoint.GetCountryCode(),
		Latitude:                   emptyToNil(checkpoint.GetLatitude()),
		Longitude:                  emptyToNil(checkpoint.GetLongitude()),
		ShippingMethod:             emptyToNil(checkpoint.GetShippingMethod()),
		OccurredAt:                 checkpoint.GetOccurredAt(),
		ReportedAt:                 checkpoint.GetReportedAt(),
		ReportedByOrganizationName: checkpoint.GetReportedByOrganizationName(),
		AnchorStatus:               anchorStatusName[checkpoint.GetAnchorStatus()],
		AnchorTransactionHash:      emptyToNil(checkpoint.GetAnchorTransactionHash()),
		InclusionProofAvailable:    checkpoint.GetInclusionProofAvailable(),
		SupersedesCheckpointID:     emptyToNil(checkpoint.GetSupersedesCheckpointId()),
		SupersededByCheckpointID:   emptyToNil(checkpoint.GetSupersededByCheckpointId()),
		CorrectionReason:           emptyToNil(checkpoint.GetCorrectionReason()),
	}

	if epoch := checkpoint.GetAnchorEpoch(); epoch != 0 {
		mapped.AnchorEpoch = &epoch
	}

	return mapped
}

func toCheckpointResponses(
	checkpoints []*provenancev1.Checkpoint,
) []checkpointResponse {
	mapped := make([]checkpointResponse, 0, len(checkpoints))
	for _, checkpoint := range checkpoints {
		mapped = append(mapped, toCheckpointResponse(checkpoint))
	}
	return mapped
}

func pageLimit(raw string) int32 {
	if raw == "" {
		return 0
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed < 0 {
		return 0
	}
	return int32(parsed)
}

func (h *Handlers) ListBatches(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	page, err := h.Provenance.ListBatches(
		c.Request.Context(), c.Query("after"), pageLimit(c.Query("per_page")),
	)
	if err != nil {
		h.failProvenance(c, err)
		return
	}

	batches := make([]batchResponse, 0, len(page.GetBatches()))
	for _, batch := range page.GetBatches() {
		batches = append(batches, toBatchResponse(batch))
	}

	httpx.Data(c, http.StatusOK, map[string]any{
		"batches":  batches,
		"cursor":   emptyToNil(page.GetCursor()),
		"has_more": page.GetHasMore(),
	})
}

func (h *Handlers) CreateBatch(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	var body createBatchRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.Fail(c, httpx.CodeValidation)
		return
	}

	category, known := batchCategoryByName[body.ProductCategory]
	if !known {
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
			Field: "product_category", Code: "UNSUPPORTED_VALUE",
		})
		return
	}

	key, present := httpx.IdempotencyKeyFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeIdempotencyKeyRequired)
		return
	}

	created, err := h.Provenance.CreateBatch(c.Request.Context(), key,
		&provenancev1.CreateBatchRequest{
			OriginatingFacilityId: body.OriginatingFacilityID,
			ProductCategory:       category,
			ComponentType:         body.ComponentType,
			LotNumber:             body.LotNumber,
			Quantity:              body.Quantity,
			Unit:                  body.Unit,
			ProducedAt:            body.ProducedAt,
			ExternalId:            body.ExternalID,
			ParentReferences:      body.ParentReferences,
		})
	if err != nil {
		h.failProvenance(c, err)
		return
	}

	httpx.Data(c, http.StatusCreated, map[string]any{
		"batch":   toBatchResponse(created.GetBatch()),
		"parents": toParentResponses(created.GetParents()),
	})
}

func (h *Handlers) GetBatch(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	view, err := h.Provenance.GetBatch(c.Request.Context(), c.Param("batchId"))
	if err != nil {
		h.failProvenance(c, err)
		return
	}

	httpx.Data(c, http.StatusOK, map[string]any{
		"batch":   toBatchResponse(view.GetBatch()),
		"parents": toParentResponses(view.GetParents()),
	})
}

func (h *Handlers) ListCheckpoints(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	page, err := h.Provenance.ListCheckpoints(
		c.Request.Context(),
		c.Param("batchId"),
		c.Query("after"),
		pageLimit(c.Query("per_page")),
	)
	if err != nil {
		h.failProvenance(c, err)
		return
	}

	httpx.Data(c, http.StatusOK, map[string]any{
		"checkpoints": toCheckpointResponses(page.GetCheckpoints()),
		"cursor":      emptyToNil(page.GetCursor()),
		"has_more":    page.GetHasMore(),
	})
}

func (h *Handlers) LogCheckpoint(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	var body logCheckpointRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.Fail(c, httpx.CodeValidation)
		return
	}

	kind, known := checkpointTypeByName[body.Type]
	if !known {
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
			Field: "type", Code: "UNSUPPORTED_VALUE",
		})
		return
	}

	key, present := httpx.IdempotencyKeyFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeIdempotencyKeyRequired)
		return
	}

	logged, err := h.Provenance.LogCheckpoint(c.Request.Context(), key,
		&provenancev1.LogCheckpointRequest{
			BatchId:              c.Param("batchId"),
			Type:                 kind,
			LocationLabel:        body.LocationLabel,
			CountryCode:          body.CountryCode,
			Latitude:             body.Latitude,
			Longitude:            body.Longitude,
			ShippingMethod:       body.ShippingMethod,
			OccurredAt:           body.OccurredAt,
			CorrectsCheckpointId: body.CorrectsCheckpointID,
			CorrectionReason:     body.CorrectionReason,
			ExternalId:           body.ExternalID,
		})
	if err != nil {
		h.failProvenance(c, err)
		return
	}

	httpx.Data(c, http.StatusCreated, map[string]any{
		"checkpoint":       toCheckpointResponse(logged.GetCheckpoint()),
		"provenance_score": toScoreResponse(logged.GetProvenanceScore()),
	})
}

func (h *Handlers) GetComponentBatch(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	view, err := h.Provenance.GetComponentBatch(
		c.Request.Context(), c.Param("batchId"), c.Param("componentBatchId"),
	)
	if err != nil {
		h.failProvenance(c, err)
		return
	}

	httpx.Data(c, http.StatusOK, map[string]any{
		"batch":       toBatchResponse(view.GetBatch()),
		"checkpoints": toCheckpointResponses(view.GetCheckpoints()),
	})
}

func (h *Handlers) failProvenance(c *gin.Context, err error) {
	switch status.Code(err) {
	case codes.PermissionDenied:
		httpx.Fail(c, httpx.CodeForbidden)
	case codes.NotFound:
		httpx.Fail(c, httpx.CodeResourceNotFound)
	case codes.InvalidArgument:
		httpx.Fail(c, httpx.CodeValidation)
	case codes.AlreadyExists, codes.Aborted, codes.FailedPrecondition:
		httpx.Fail(c, httpx.CodeConflict)
	default:
		h.Logger.Error("provenance upstream failed", errorAttributes(c, err)...)
		httpx.Fail(c, httpx.CodeDependencyUnavailable)
	}
}
