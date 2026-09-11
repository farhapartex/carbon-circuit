package rpc

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	creditledgerv1 "github.com/carboncircuit/backend/gen/carboncircuit/creditledger/v1"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/services/credit-ledger-service/internal/domain"
	"github.com/carboncircuit/backend/services/credit-ledger-service/internal/service"
)

type PortfolioReader interface {
	Portfolio(ctx context.Context, actor service.Actor) (service.Portfolio, error)
}

type CreditLedgerServer struct {
	creditledgerv1.UnimplementedCreditLedgerServiceServer

	database *gorm.DB
	reader   PortfolioReader
	logger   *slog.Logger
	revision string
}

func NewCreditLedgerServer(
	database *gorm.DB,
	reader PortfolioReader,
	logger *slog.Logger,
	revision string,
) *CreditLedgerServer {
	return &CreditLedgerServer{
		database: database,
		reader:   reader,
		logger:   logger,
		revision: revision,
	}
}

func (s *CreditLedgerServer) Ping(
	ctx context.Context,
	_ *creditledgerv1.PingRequest,
) (*creditledgerv1.PingResponse, error) {
	reachable := s.databaseReachable(ctx)

	var classes int64
	if reachable {
		s.database.Table("credit_ledger.credit_classes").
			Where("deleted_at IS NULL").Count(&classes)
	}

	return &creditledgerv1.PingResponse{
		Service:           "credit-ledger-service",
		Revision:          s.revision,
		DatabaseReachable: reachable,
		CreditClassCount:  int32(classes),
	}, nil
}

func (s *CreditLedgerServer) databaseReachable(ctx context.Context) bool {
	pool, err := s.database.DB()
	if err != nil {
		return false
	}
	return pool.PingContext(ctx) == nil
}

func (s *CreditLedgerServer) GetPortfolio(
	ctx context.Context,
	_ *creditledgerv1.GetPortfolioRequest,
) (*creditledgerv1.GetPortfolioResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	portfolio, err := s.reader.Portfolio(ctx, actor)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	holdings := make([]*creditledgerv1.Holding, 0, len(portfolio.Holdings))
	for _, holding := range portfolio.Holdings {
		holdings = append(holdings, &creditledgerv1.Holding{
			CreditClass: classMessage(holding.Class),
			Available:   holding.Available,
			Escrowed:    holding.Escrowed,
			Retired:     holding.Retired,
		})
	}

	return &creditledgerv1.GetPortfolioResponse{
		Holdings:       holdings,
		TotalAvailable: portfolio.TotalAvailable,
		TotalEscrowed:  portfolio.TotalEscrowed,
		TotalRetired:   portfolio.TotalRetired,
	}, nil
}

func classMessage(class domain.CreditClass) *creditledgerv1.CreditClass {
	return &creditledgerv1.CreditClass{
		TokenId:         class.TokenID,
		FacilityId:      class.FacilityID.String(),
		FacilityName:    class.FacilityName,
		FacilityCountry: class.FacilityCountry,
		VintageYear:     int32(class.VintageYear),
		ActivityType:    string(class.ActivityType),
	}
}

func (s *CreditLedgerServer) actor(ctx context.Context) (service.Actor, error) {
	verified, present := grpcx.CallerFrom(ctx)
	if !present || !verified.HasOrganization() {
		return service.Actor{}, status.Error(codes.Unauthenticated, "a verified organization is required")
	}

	organizationID, err := uuid.Parse(verified.OrganizationID)
	if err != nil {
		return service.Actor{}, status.Error(codes.Unauthenticated, "service token carries an unusable organization")
	}

	userID, err := uuid.Parse(verified.UserID)
	if err != nil {
		return service.Actor{}, status.Error(codes.Unauthenticated, "service token carries an unusable user")
	}

	return service.Actor{OrganizationID: organizationID, UserID: userID}, nil
}
