package rpc

import (
	"context"

	"gorm.io/gorm"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
)

type IdentityServer struct {
	identityv1.UnimplementedIdentityServiceServer

	database *gorm.DB
	revision string
}

func NewIdentityServer(database *gorm.DB, revision string) *IdentityServer {
	return &IdentityServer{database: database, revision: revision}
}

func (s *IdentityServer) Ping(
	ctx context.Context,
	_ *identityv1.PingRequest,
) (*identityv1.PingResponse, error) {
	return &identityv1.PingResponse{
		Service:           "identity-service",
		Revision:          s.revision,
		DatabaseReachable: s.DatabaseReachable(ctx),
	}, nil
}

func (s *IdentityServer) DatabaseReachable(ctx context.Context) bool {
	pool, err := s.database.DB()
	if err != nil {
		return false
	}
	return pool.PingContext(ctx) == nil
}
