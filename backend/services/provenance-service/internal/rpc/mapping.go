package rpc

import (
	"encoding/json"
	"time"

	provenancev1 "github.com/carboncircuit/backend/gen/carboncircuit/provenance/v1"
	"github.com/carboncircuit/backend/services/provenance-service/internal/domain"
	"github.com/carboncircuit/backend/services/provenance-service/internal/score"
	"github.com/carboncircuit/backend/services/provenance-service/internal/service"
)

var categoriesToProto = map[domain.ProductCategory]provenancev1.ProductCategory{
	domain.CategoryElectronics: provenancev1.ProductCategory_PRODUCT_CATEGORY_ELECTRONICS,
	domain.CategoryAgriculture: provenancev1.ProductCategory_PRODUCT_CATEGORY_AGRICULTURE,
	domain.CategoryPharma:      provenancev1.ProductCategory_PRODUCT_CATEGORY_PHARMA,
	domain.CategoryTextiles:    provenancev1.ProductCategory_PRODUCT_CATEGORY_TEXTILES,
}

var categoriesFromProto = map[provenancev1.ProductCategory]domain.ProductCategory{
	provenancev1.ProductCategory_PRODUCT_CATEGORY_ELECTRONICS: domain.CategoryElectronics,
	provenancev1.ProductCategory_PRODUCT_CATEGORY_AGRICULTURE: domain.CategoryAgriculture,
	provenancev1.ProductCategory_PRODUCT_CATEGORY_PHARMA:      domain.CategoryPharma,
	provenancev1.ProductCategory_PRODUCT_CATEGORY_TEXTILES:    domain.CategoryTextiles,
}

var checkpointTypesToProto = map[domain.CheckpointType]provenancev1.CheckpointType{
	domain.ProductionComplete: provenancev1.CheckpointType_CHECKPOINT_TYPE_PRODUCTION_COMPLETE,
	domain.DepartedOrigin:     provenancev1.CheckpointType_CHECKPOINT_TYPE_DEPARTED_ORIGIN,
	domain.CustomsExport:      provenancev1.CheckpointType_CHECKPOINT_TYPE_CUSTOMS_EXPORT,
	domain.CustomsImport:      provenancev1.CheckpointType_CHECKPOINT_TYPE_CUSTOMS_IMPORT,
	domain.ArrivedDestination: provenancev1.CheckpointType_CHECKPOINT_TYPE_ARRIVED_DESTINATION,
}

var checkpointTypesFromProto = map[provenancev1.CheckpointType]domain.CheckpointType{
	provenancev1.CheckpointType_CHECKPOINT_TYPE_PRODUCTION_COMPLETE: domain.ProductionComplete,
	provenancev1.CheckpointType_CHECKPOINT_TYPE_DEPARTED_ORIGIN:     domain.DepartedOrigin,
	provenancev1.CheckpointType_CHECKPOINT_TYPE_CUSTOMS_EXPORT:      domain.CustomsExport,
	provenancev1.CheckpointType_CHECKPOINT_TYPE_CUSTOMS_IMPORT:      domain.CustomsImport,
	provenancev1.CheckpointType_CHECKPOINT_TYPE_ARRIVED_DESTINATION: domain.ArrivedDestination,
}

var anchorStatusToProto = map[domain.AnchorStatus]provenancev1.AnchorStatus{
	domain.Unanchored:  provenancev1.AnchorStatus_ANCHOR_STATUS_UNANCHORED,
	domain.Provisional: provenancev1.AnchorStatus_ANCHOR_STATUS_PROVISIONAL,
	domain.Confirmed:   provenancev1.AnchorStatus_ANCHOR_STATUS_CONFIRMED,
}

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func intOrZero(value *int) int32 {
	if value == nil {
		return 0
	}
	return int32(*value)
}

func scoreToProto(batch domain.Batch) *provenancev1.ProvenanceScore {
	var components []score.Component
	if len(batch.ScoreComponents) > 0 {
		_ = json.Unmarshal(batch.ScoreComponents, &components)
	}

	mapped := make([]*provenancev1.ScoreComponent, 0, len(components))
	for _, component := range components {
		mapped = append(mapped, &provenancev1.ScoreComponent{
			Label:       component.Label,
			Earned:      int32(component.Earned),
			Available:   int32(component.Available),
			Explanation: component.Explanation,
		})
	}

	return &provenancev1.ProvenanceScore{
		Total:      int32(batch.ProvenanceScore),
		Components: mapped,
	}
}

func batchToProto(batch domain.Batch) *provenancev1.Batch {
	return &provenancev1.Batch{
		Id:                      batch.ID.String(),
		OrganizationId:          batch.OrganizationID.String(),
		OriginatingFacilityId:   batch.OriginatingFacilityID.String(),
		OriginatingFacilityName: batch.OriginatingFacilityName,
		PublicReference:         batch.PublicReference,
		ProductCategory:         categoriesToProto[batch.ProductCategory],
		ComponentType:           batch.ComponentType,
		LotNumber:               stringOrEmpty(batch.LotNumber),
		Quantity:                batch.Quantity,
		Unit:                    batch.Unit,
		ProducedAt:              batch.ProducedAt.UTC().Format(time.RFC3339),
		ExternalId:              stringOrEmpty(batch.ExternalID),
		CheckpointCount:         int32(batch.CheckpointCount),
		ProvenanceScore:         scoreToProto(batch),
		CreatedAt:               batch.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func parentsToProto(parents []service.ParentView) []*provenancev1.ParentBatch {
	mapped := make([]*provenancev1.ParentBatch, 0, len(parents))

	for _, parent := range parents {
		entry := &provenancev1.ParentBatch{
			DeclaredReference: parent.DeclaredReference,
			Resolved:          parent.Resolved,
		}

		if parent.Batch != nil {
			entry.Id = parent.Batch.ID.String()
			entry.ComponentType = parent.Batch.ComponentType
			entry.ProductCategory = categoriesToProto[parent.Batch.ProductCategory]
			entry.OriginatingFacilityName = parent.Batch.OriginatingFacilityName
		}

		mapped = append(mapped, entry)
	}

	return mapped
}

func checkpointToProto(checkpoint domain.Checkpoint) *provenancev1.Checkpoint {
	mapped := &provenancev1.Checkpoint{
		Id:                         checkpoint.ID.String(),
		BatchId:                    checkpoint.BatchID.String(),
		Type:                       checkpointTypesToProto[checkpoint.Type],
		LocationLabel:              checkpoint.LocationLabel,
		CountryCode:                checkpoint.CountryCode,
		Latitude:                   stringOrEmpty(checkpoint.Latitude),
		Longitude:                  stringOrEmpty(checkpoint.Longitude),
		ShippingMethod:             stringOrEmpty(checkpoint.ShippingMethod),
		OccurredAt:                 checkpoint.OccurredAt.UTC().Format(time.RFC3339),
		ReportedAt:                 checkpoint.ReportedAt.UTC().Format(time.RFC3339),
		ReportedByOrganizationName: checkpoint.ReportedByOrganizationName,
		AnchorStatus:               anchorStatusToProto[checkpoint.AnchorStatus],
		AnchorEpoch:                intOrZero(checkpoint.AnchorEpoch),
		AnchorTransactionHash:      stringOrEmpty(checkpoint.AnchorTransactionHash),
		InclusionProofAvailable:    checkpoint.InclusionProofAvailable,
		CorrectionReason:           stringOrEmpty(checkpoint.CorrectionReason),
	}

	if checkpoint.SupersedesCheckpointID != nil {
		mapped.SupersedesCheckpointId = checkpoint.SupersedesCheckpointID.String()
	}
	if checkpoint.SupersededByCheckpointID != nil {
		mapped.SupersededByCheckpointId = checkpoint.SupersededByCheckpointID.String()
	}

	return mapped
}

func checkpointsToProto(checkpoints []domain.Checkpoint) []*provenancev1.Checkpoint {
	mapped := make([]*provenancev1.Checkpoint, 0, len(checkpoints))
	for _, checkpoint := range checkpoints {
		mapped = append(mapped, checkpointToProto(checkpoint))
	}
	return mapped
}
