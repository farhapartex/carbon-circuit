package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	provenancereadv1 "github.com/carboncircuit/backend/gen/carboncircuit/provenanceread/v1"
	"github.com/carboncircuit/backend/internal/httpx"
)

type publicCheckpointResponse struct {
	Type                  string  `json:"type"`
	LocationLabel         string  `json:"location_label"`
	CountryCode           string  `json:"country_code"`
	ShippingMethod        *string `json:"shipping_method"`
	OccurredAt            string  `json:"occurred_at"`
	AnchorStatus          string  `json:"anchor_status"`
	AnchorEpoch           *int32  `json:"anchor_epoch"`
	AnchorTransactionHash *string `json:"anchor_transaction_hash"`
	Superseded            bool    `json:"superseded"`
}

type publicBatchResponse struct {
	PublicReference            string                     `json:"public_reference"`
	ProductCategory            string                     `json:"product_category"`
	ComponentType              string                     `json:"component_type"`
	OriginatingFacilityName    string                     `json:"originating_facility_name"`
	OriginatingFacilityCountry string                     `json:"originating_facility_country"`
	ProducedAt                 string                     `json:"produced_at"`
	ProvenanceScore            provenanceScoreResponse    `json:"provenance_score"`
	Checkpoints                []publicCheckpointResponse `json:"checkpoints"`
	LastUpdatedAt              string                     `json:"last_updated_at"`
}

func (h *Handlers) TrackBatch(c *gin.Context) {
	tracked, err := h.ProvenanceRead.TrackBatch(
		c.Request.Context(), c.Param("publicRef"),
	)
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			httpx.Fail(c, httpx.CodeResourceNotFound)
		case codes.InvalidArgument:
			httpx.Fail(c, httpx.CodeValidation)
		default:
			h.Logger.Error("provenance read upstream failed", errorAttributes(c, err)...)
			httpx.Fail(c, httpx.CodeDependencyUnavailable)
		}
		return
	}

	httpx.Data(c, http.StatusOK, toPublicBatchResponse(tracked))
}

func toPublicBatchResponse(
	tracked *provenancereadv1.TrackBatchResponse,
) publicBatchResponse {
	components := make([]scoreComponentResponse, 0, len(tracked.GetScoreComponents()))
	for _, component := range tracked.GetScoreComponents() {
		components = append(components, scoreComponentResponse{
			Label:       component.GetLabel(),
			Earned:      component.GetEarned(),
			Available:   component.GetAvailable(),
			Explanation: component.GetExplanation(),
		})
	}

	checkpoints := make([]publicCheckpointResponse, 0, len(tracked.GetCheckpoints()))
	for _, checkpoint := range tracked.GetCheckpoints() {
		entry := publicCheckpointResponse{
			Type:                  checkpoint.GetType(),
			LocationLabel:         checkpoint.GetLocationLabel(),
			CountryCode:           checkpoint.GetCountryCode(),
			ShippingMethod:        emptyToNil(checkpoint.GetShippingMethod()),
			OccurredAt:            checkpoint.GetOccurredAt(),
			AnchorStatus:          checkpoint.GetAnchorStatus(),
			AnchorTransactionHash: emptyToNil(checkpoint.GetAnchorTransactionHash()),
			Superseded:            checkpoint.GetSuperseded(),
		}
		if epoch := checkpoint.GetAnchorEpoch(); epoch != 0 {
			entry.AnchorEpoch = &epoch
		}
		checkpoints = append(checkpoints, entry)
	}

	return publicBatchResponse{
		PublicReference:            tracked.GetPublicReference(),
		ProductCategory:            tracked.GetProductCategory(),
		ComponentType:              tracked.GetComponentType(),
		OriginatingFacilityName:    tracked.GetOriginatingFacilityName(),
		OriginatingFacilityCountry: tracked.GetOriginatingFacilityCountry(),
		ProducedAt:                 tracked.GetProducedAt(),
		ProvenanceScore: provenanceScoreResponse{
			Total:      tracked.GetProvenanceScore(),
			Components: components,
		},
		Checkpoints:   checkpoints,
		LastUpdatedAt: tracked.GetLastUpdatedAt(),
	}
}
