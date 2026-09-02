package rpc

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	provenancev1 "github.com/carboncircuit/backend/gen/carboncircuit/provenance/v1"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/servicetoken"
	"github.com/carboncircuit/backend/services/provenance-service/internal/repository"
	"github.com/carboncircuit/backend/services/provenance-service/internal/service"
)

type BatchManager interface {
	Create(ctx context.Context, actor service.Actor, declaration service.BatchDeclaration) (service.BatchView, error)
	List(ctx context.Context, actor service.Actor, after string, limit int) (service.BatchPage, error)
	Get(ctx context.Context, actor service.Actor, batchID uuid.UUID) (service.BatchView, error)
	Checkpoints(ctx context.Context, actor service.Actor, batchID uuid.UUID, after string, limit int) (service.CheckpointPage, error)
	Component(ctx context.Context, actor service.Actor, batchID, componentBatchID uuid.UUID) (service.ComponentView, error)
	LogCheckpoint(ctx context.Context, actor service.Actor, batchID uuid.UUID, entry service.CheckpointEntry) (service.LoggedCheckpoint, error)
}

type ProvenanceServer struct {
	provenancev1.UnimplementedProvenanceServiceServer

	database *gorm.DB
	batches  BatchManager
	logger   *slog.Logger
	revision string
}

func NewProvenanceServer(
	database *gorm.DB,
	batches BatchManager,
	logger *slog.Logger,
	revision string,
) *ProvenanceServer {
	return &ProvenanceServer{
		database: database,
		batches:  batches,
		logger:   logger,
		revision: revision,
	}
}

func (s *ProvenanceServer) Ping(
	ctx context.Context,
	_ *provenancev1.PingRequest,
) (*provenancev1.PingResponse, error) {
	return &provenancev1.PingResponse{
		Service:           "provenance-service",
		Revision:          s.revision,
		DatabaseReachable: s.databaseReachable(ctx),
	}, nil
}

func (s *ProvenanceServer) databaseReachable(ctx context.Context) bool {
	pool, err := s.database.DB()
	if err != nil {
		return false
	}
	return pool.PingContext(ctx) == nil
}

func (s *ProvenanceServer) actor(ctx context.Context) (service.Actor, error) {
	verified, present := grpcx.CallerFrom(ctx)
	if !present || !verified.HasOrganization() {
		return service.Actor{}, status.Error(codes.Unauthenticated, "a verified organization is required")
	}
	return actorFrom(verified)
}

func actorFrom(verified servicetoken.Caller) (service.Actor, error) {
	organizationID, err := uuid.Parse(verified.OrganizationID)
	if err != nil {
		return service.Actor{}, status.Error(codes.Unauthenticated, "service token carries an unusable organization")
	}

	userID, err := uuid.Parse(verified.UserID)
	if err != nil {
		return service.Actor{}, status.Error(codes.Unauthenticated, "service token carries an unusable user")
	}

	return service.Actor{
		OrganizationID:    organizationID,
		UserID:            userID,
		OrganizationName:  verified.OrganizationName,
		OrganizationType:  verified.OrganizationType,
		PlanTier:          verified.PlanTier,
		OrganizationState: verified.OrganizationState,
	}, nil
}

func (s *ProvenanceServer) CreateBatch(
	ctx context.Context,
	request *provenancev1.CreateBatchRequest,
) (*provenancev1.CreateBatchResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	idempotencyKey := grpcx.IdempotencyKeyFromIncoming(ctx)
	if idempotencyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "an idempotency key is required")
	}

	facilityID, err := uuid.Parse(request.GetOriginatingFacilityId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "originating facility id is not a valid identifier")
	}

	category, known := categoriesFromProto[request.GetProductCategory()]
	if !known {
		return nil, status.Error(codes.InvalidArgument, "product category must be a known category")
	}

	producedAt, err := time.Parse(time.RFC3339, request.GetProducedAt())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "production date must be an RFC3339 timestamp")
	}

	created, err := s.batches.Create(ctx, actor, service.BatchDeclaration{
		OriginatingFacilityID: facilityID,
		ProductCategory:       category,
		ComponentType:         request.GetComponentType(),
		LotNumber:             request.GetLotNumber(),
		Quantity:              request.GetQuantity(),
		Unit:                  request.GetUnit(),
		ProducedAt:            producedAt,
		ExternalID:            request.GetExternalId(),
		ParentReferences:      request.GetParentReferences(),
		IdempotencyKey:        idempotencyKey,
		RequestBody:           canonicalBatchRequest(request),
	})
	if err != nil {
		return nil, provenanceFailure(err)
	}

	return &provenancev1.CreateBatchResponse{
		Batch:    batchToProto(created.Batch),
		Parents:  parentsToProto(created.Parents),
		Replayed: created.Replayed,
	}, nil
}

func (s *ProvenanceServer) ListBatches(
	ctx context.Context,
	request *provenancev1.ListBatchesRequest,
) (*provenancev1.ListBatchesResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	page, err := s.batches.List(ctx, actor, request.GetAfter(), int(request.GetLimit()))
	if err != nil {
		return nil, provenanceFailure(err)
	}

	batches := make([]*provenancev1.Batch, 0, len(page.Batches))
	for _, batch := range page.Batches {
		batches = append(batches, batchToProto(batch))
	}

	return &provenancev1.ListBatchesResponse{
		Batches: batches,
		Cursor:  page.Cursor,
		HasMore: page.HasMore,
	}, nil
}

func (s *ProvenanceServer) GetBatch(
	ctx context.Context,
	request *provenancev1.GetBatchRequest,
) (*provenancev1.GetBatchResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	batchID, err := uuid.Parse(request.GetBatchId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "batch id is not a valid identifier")
	}

	view, err := s.batches.Get(ctx, actor, batchID)
	if err != nil {
		return nil, provenanceFailure(err)
	}

	return &provenancev1.GetBatchResponse{
		Batch:   batchToProto(view.Batch),
		Parents: parentsToProto(view.Parents),
	}, nil
}

func (s *ProvenanceServer) ListCheckpoints(
	ctx context.Context,
	request *provenancev1.ListCheckpointsRequest,
) (*provenancev1.ListCheckpointsResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	batchID, err := uuid.Parse(request.GetBatchId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "batch id is not a valid identifier")
	}

	page, err := s.batches.Checkpoints(
		ctx, actor, batchID, request.GetAfter(), int(request.GetLimit()),
	)
	if err != nil {
		return nil, provenanceFailure(err)
	}

	return &provenancev1.ListCheckpointsResponse{
		Checkpoints: checkpointsToProto(page.Checkpoints),
		Cursor:      page.Cursor,
		HasMore:     page.HasMore,
	}, nil
}

func (s *ProvenanceServer) LogCheckpoint(
	ctx context.Context,
	request *provenancev1.LogCheckpointRequest,
) (*provenancev1.LogCheckpointResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	idempotencyKey := grpcx.IdempotencyKeyFromIncoming(ctx)
	if idempotencyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "an idempotency key is required")
	}

	batchID, err := uuid.Parse(request.GetBatchId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "batch id is not a valid identifier")
	}

	kind, known := checkpointTypesFromProto[request.GetType()]
	if !known {
		return nil, status.Error(codes.InvalidArgument, "checkpoint type must be a known type")
	}

	occurredAt, err := time.Parse(time.RFC3339, request.GetOccurredAt())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "event time must be an RFC3339 timestamp")
	}

	entry := service.CheckpointEntry{
		Type:             kind,
		LocationLabel:    request.GetLocationLabel(),
		CountryCode:      request.GetCountryCode(),
		Latitude:         request.GetLatitude(),
		Longitude:        request.GetLongitude(),
		ShippingMethod:   request.GetShippingMethod(),
		OccurredAt:       occurredAt,
		CorrectionReason: request.GetCorrectionReason(),
		ExternalID:       request.GetExternalId(),
		IdempotencyKey:   idempotencyKey,
		RequestBody:      canonicalCheckpointRequest(request),
	}

	if corrects := request.GetCorrectsCheckpointId(); corrects != "" {
		parsed, parseErr := uuid.Parse(corrects)
		if parseErr != nil {
			return nil, status.Error(codes.InvalidArgument, "corrected checkpoint id is not a valid identifier")
		}
		entry.CorrectsID = &parsed
	}

	logged, err := s.batches.LogCheckpoint(ctx, actor, batchID, entry)
	if err != nil {
		return nil, provenanceFailure(err)
	}

	return &provenancev1.LogCheckpointResponse{
		Checkpoint:      checkpointToProto(logged.Checkpoint),
		ProvenanceScore: scoreToProto(logged.Batch),
		Replayed:        logged.Replayed,
	}, nil
}

func (s *ProvenanceServer) GetComponentBatch(
	ctx context.Context,
	request *provenancev1.GetComponentBatchRequest,
) (*provenancev1.GetComponentBatchResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	batchID, err := uuid.Parse(request.GetBatchId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "batch id is not a valid identifier")
	}

	componentID, err := uuid.Parse(request.GetComponentBatchId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "component batch id is not a valid identifier")
	}

	view, err := s.batches.Component(ctx, actor, batchID, componentID)
	if err != nil {
		return nil, provenanceFailure(err)
	}

	return &provenancev1.GetComponentBatchResponse{
		Batch:       batchToProto(view.Batch),
		Checkpoints: checkpointsToProto(view.Checkpoints),
	}, nil
}

func canonicalBatchRequest(request *provenancev1.CreateBatchRequest) []byte {
	parts := []string{
		request.GetOriginatingFacilityId(),
		request.GetProductCategory().String(),
		strings.TrimSpace(request.GetComponentType()),
		strings.TrimSpace(request.GetLotNumber()),
		strings.TrimSpace(request.GetQuantity()),
		strings.TrimSpace(request.GetUnit()),
		request.GetProducedAt(),
		strings.TrimSpace(request.GetExternalId()),
		strings.Join(request.GetParentReferences(), ","),
	}
	return []byte(strings.Join(parts, "\x1f"))
}

func canonicalCheckpointRequest(request *provenancev1.LogCheckpointRequest) []byte {
	parts := []string{
		request.GetBatchId(),
		request.GetType().String(),
		strings.TrimSpace(request.GetLocationLabel()),
		strings.ToUpper(request.GetCountryCode()),
		request.GetLatitude(),
		request.GetLongitude(),
		request.GetShippingMethod(),
		request.GetOccurredAt(),
		request.GetCorrectsCheckpointId(),
		strings.TrimSpace(request.GetCorrectionReason()),
		strings.TrimSpace(request.GetExternalId()),
	}
	return []byte(strings.Join(parts, "\x1f"))
}

func provenanceFailure(err error) error {
	switch {
	case errors.Is(err, service.ErrBatchNotFound), errors.Is(err, service.ErrCheckpointNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrNotBatchOwner):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, service.ErrOrganizationReadOnly), errors.Is(err, service.ErrPlanForbidsBatches):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, service.ErrFacilityUnknown):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, service.ErrRequestInProgress):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, service.ErrIdempotencyConflict), errors.Is(err, service.ErrExternalIDTaken):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, service.ErrAlreadyCorrected), errors.Is(err, service.ErrCorrectionWindowClosed):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, service.ErrChainTooDeep),
		errors.Is(err, service.ErrCheckpointBeforeProduction),
		errors.Is(err, service.ErrCheckpointInFuture),
		errors.Is(err, service.ErrShippingMethodRequired),
		errors.Is(err, service.ErrCorrectionReasonRequired):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, repository.ErrConcurrentUpdate):
		return status.Error(codes.Aborted, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
