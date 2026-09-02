package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	provenancereadv1 "github.com/carboncircuit/backend/gen/carboncircuit/provenanceread/v1"
	"github.com/carboncircuit/backend/internal/events"
	"github.com/carboncircuit/backend/services/provenance-read-service/internal/domain"
	"github.com/carboncircuit/backend/services/provenance-read-service/internal/service"
)

type PublicTracker interface {
	Track(ctx context.Context, reference string) (service.PublicView, error)
}

type ProvenanceReadServer struct {
	provenancereadv1.UnimplementedProvenanceReadServiceServer

	database *gorm.DB
	tracker  PublicTracker
	logger   *slog.Logger
	revision string
}

func NewProvenanceReadServer(
	database *gorm.DB,
	tracker PublicTracker,
	logger *slog.Logger,
	revision string,
) *ProvenanceReadServer {
	return &ProvenanceReadServer{
		database: database,
		tracker:  tracker,
		logger:   logger,
		revision: revision,
	}
}

func (s *ProvenanceReadServer) Ping(
	ctx context.Context,
	_ *provenancereadv1.PingRequest,
) (*provenancereadv1.PingResponse, error) {
	pool, err := s.database.DB()
	reachable := err == nil && pool.PingContext(ctx) == nil

	return &provenancereadv1.PingResponse{
		Service:           "provenance-read-service",
		Revision:          s.revision,
		DatabaseReachable: reachable,
	}, nil
}

func (s *ProvenanceReadServer) TrackBatch(
	ctx context.Context,
	request *provenancereadv1.TrackBatchRequest,
) (*provenancereadv1.TrackBatchResponse, error) {
	reference := request.GetPublicReference()
	if len(reference) != 22 {
		return nil, status.Error(codes.InvalidArgument, "a public batch reference is 22 characters")
	}

	view, err := s.tracker.Track(ctx, reference)
	if err != nil {
		if errors.Is(err, service.ErrBatchNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &provenancereadv1.TrackBatchResponse{
		PublicReference:            view.Batch.PublicReference,
		ProductCategory:            view.Batch.ProductCategory,
		ComponentType:              view.Batch.ComponentType,
		OriginatingFacilityName:    view.Batch.OriginatingFacilityName,
		OriginatingFacilityCountry: view.Batch.OriginatingFacilityCountry,
		ProducedAt:                 view.Batch.ProducedAt.UTC().Format(time.RFC3339),
		ProvenanceScore:            int32(view.Batch.ProvenanceScore),
		ScoreComponents:            componentsToProto(view.Batch),
		Checkpoints:                checkpointsToProto(view.Checkpoints),
		LastUpdatedAt:              view.Batch.LastUpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func componentsToProto(batch domain.PublicBatch) []*provenancereadv1.ScoreComponent {
	var decoded []events.ScoreComponent
	if len(batch.ScoreComponents) > 0 {
		_ = json.Unmarshal(batch.ScoreComponents, &decoded)
	}

	mapped := make([]*provenancereadv1.ScoreComponent, 0, len(decoded))
	for _, component := range decoded {
		mapped = append(mapped, &provenancereadv1.ScoreComponent{
			Label:       component.Label,
			Earned:      int32(component.Earned),
			Available:   int32(component.Available),
			Explanation: component.Explanation,
		})
	}

	return mapped
}

func checkpointsToProto(
	checkpoints []domain.PublicCheckpoint,
) []*provenancereadv1.PublicCheckpoint {
	mapped := make([]*provenancereadv1.PublicCheckpoint, 0, len(checkpoints))

	for _, checkpoint := range checkpoints {
		entry := &provenancereadv1.PublicCheckpoint{
			Type:          checkpoint.Type,
			LocationLabel: checkpoint.LocationLabel,
			CountryCode:   checkpoint.CountryCode,
			OccurredAt:    checkpoint.OccurredAt.UTC().Format(time.RFC3339),
			AnchorStatus:  checkpoint.AnchorStatus,
			Superseded:    checkpoint.Superseded,
		}

		if checkpoint.ShippingMethod != nil {
			entry.ShippingMethod = *checkpoint.ShippingMethod
		}
		if checkpoint.AnchorEpoch != nil {
			entry.AnchorEpoch = int32(*checkpoint.AnchorEpoch)
		}
		if checkpoint.AnchorTransactionHash != nil {
			entry.AnchorTransactionHash = *checkpoint.AnchorTransactionHash
		}

		mapped = append(mapped, entry)
	}

	return mapped
}
