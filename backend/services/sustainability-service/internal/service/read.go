package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/ceiling"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/domain"
)

type ClaimPage struct {
	Claims     []domain.Claim
	NextCursor string
}

func (s *ClaimService) Get(
	ctx context.Context,
	actor Actor,
	claimID uuid.UUID,
) (ClaimView, error) {
	var view ClaimView

	err := database.WithinTenant(ctx, s.database, tenancy(actor), func(tx database.Tx) error {
		claim, found, err := s.claims.Find(tx, actor.OrganizationID, claimID)
		if err != nil {
			return err
		}
		if !found {
			return ErrClaimNotFound
		}

		attachments, err := s.claims.Evidence(tx, actor.OrganizationID, claimID)
		if err != nil {
			return err
		}

		view = ClaimView{Claim: claim, Evidence: attachments}
		return nil
	})
	if err != nil {
		return ClaimView{}, err
	}

	return view, nil
}

func (s *ClaimService) List(
	ctx context.Context,
	actor Actor,
	status domain.ClaimStatus,
	after string,
	limit int,
) (ClaimPage, error) {
	if limit <= 0 {
		limit = defaultPageSize
	}
	if limit > maximumPageSize {
		limit = maximumPageSize
	}

	var page ClaimPage

	err := database.WithinTenant(ctx, s.database, tenancy(actor), func(tx database.Tx) error {
		claims, err := s.claims.List(tx, actor.OrganizationID, status, after, limit+1)
		if err != nil {
			return err
		}

		if len(claims) > limit {
			page.NextCursor = claims[limit-1].ID.String()
			claims = claims[:limit]
		}

		page.Claims = claims
		return nil
	})
	if err != nil {
		return ClaimPage{}, err
	}

	return page, nil
}

type CeilingPreview struct {
	Ceiling        string
	VintageCeiling string
	Consumed       string
	Remaining      string
	PeriodCeiling  string
	CapacityBasis  string
	CapacitySource string
	DiscountFactor string
	ReferenceValue string
	GridRegion     string
	PeriodDays     int
	VintageDays    int
}

func (s *ClaimService) PreviewCeiling(
	ctx context.Context,
	actor Actor,
	submission Submission,
) (CeilingPreview, error) {
	if err := actor.maySubmit(); err != nil {
		return CeilingPreview{}, err
	}
	if submission.ActivityType != domain.RenewableEnergy {
		return CeilingPreview{}, ErrActivityUnsupported
	}

	facility, err := s.facilities.Facility(ctx, submission.FacilityID)
	if err != nil {
		return CeilingPreview{}, err
	}

	var preview CeilingPreview

	err = database.WithinTenant(ctx, s.database, tenancy(actor), func(tx database.Tx) error {
		computed, err := s.computeCeiling(tx, actor, facility, submission)
		if err != nil {
			return err
		}

		preview = CeilingPreview{
			Ceiling:        computed.Allowance.Effective.StringFixed(ceiling.Places),
			VintageCeiling: computed.Allowance.VintageCeiling.StringFixed(ceiling.Places),
			Consumed:       computed.Allowance.Consumed.StringFixed(ceiling.Places),
			Remaining:      computed.Allowance.Remaining.StringFixed(ceiling.Places),
			PeriodCeiling:  computed.Allowance.PeriodCeiling.StringFixed(ceiling.Places),
			CapacityBasis:  computed.Capacity.Value.String(),
			CapacitySource: computed.Capacity.Source,
			DiscountFactor: computed.Discount.StringFixed(2),
			ReferenceValue: computed.Factor.Factor,
			GridRegion:     facility.GridRegion,
			PeriodDays:     computed.Allowance.PeriodDays,
			VintageDays:    computed.Allowance.VintageDays,
		}
		return nil
	})
	if err != nil {
		return CeilingPreview{}, err
	}

	return preview, nil
}
