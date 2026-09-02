package caller

import (
	"context"
	"fmt"
	"time"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/internal/auth"
	"github.com/carboncircuit/backend/internal/cache"
	"github.com/carboncircuit/backend/internal/servicetoken"
)

const cacheKeyPrefix = "caller:context:v1:"

type SessionResolver interface {
	ResolveSession(ctx context.Context, caller auth.Caller) (*identityv1.ResolveSessionResponse, error)
}

type Resolver struct {
	cache    *cache.Client
	identity SessionResolver
	ttl      time.Duration
}

func NewResolver(client *cache.Client, identity SessionResolver, ttl time.Duration) *Resolver {
	return &Resolver{cache: client, identity: identity, ttl: ttl}
}

func cacheKey(subject string) string { return cacheKeyPrefix + subject }

func (r *Resolver) Resolve(
	ctx context.Context,
	verified auth.Caller,
) (servicetoken.Caller, error) {
	return cache.ReadThrough(ctx, r.cache, cacheKey(verified.Subject), r.ttl,
		func(loadCtx context.Context) (servicetoken.Caller, error) {
			resolved, err := r.identity.ResolveSession(loadCtx, verified)
			if err != nil {
				return servicetoken.Caller{}, fmt.Errorf("resolve caller context: %w", err)
			}
			return contextFrom(verified, resolved), nil
		},
	)
}

func (r *Resolver) Invalidate(ctx context.Context, subject string) {
	r.cache.Invalidate(ctx, cacheKey(subject))
}

func contextFrom(
	verified auth.Caller,
	resolved *identityv1.ResolveSessionResponse,
) servicetoken.Caller {
	context := servicetoken.Caller{
		Subject:      verified.Subject,
		UserID:       resolved.GetUser().GetId(),
		PlatformRole: platformRoleName[resolved.GetUser().GetPlatformRole()],
	}

	organization := resolved.GetOrganization()
	if organization == nil {
		return context
	}

	context.OrganizationID = organization.GetId()
	context.OrganizationType = organizationTypeName[organization.GetType()]
	context.OrganizationState = organizationStateName[organization.GetState()]
	context.VerificationStatus = verificationStatusName[organization.GetVerificationStatus()]
	context.Role = organizationRoleName[resolved.GetRole()]

	return context
}
