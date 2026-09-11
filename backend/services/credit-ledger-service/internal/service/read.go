package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/credit-ledger-service/internal/domain"
	"github.com/carboncircuit/backend/services/credit-ledger-service/internal/repository"
)

const maximumIssuances = 100

type Portfolio struct {
	Holdings       []domain.Holding
	TotalAvailable string
	TotalEscrowed  string
	TotalRetired   string
}

type Actor struct {
	OrganizationID uuid.UUID
	UserID         uuid.UUID
}

func (a Actor) tenancy() database.TenantContext {
	return database.TenantContext{
		UserID:         a.UserID.String(),
		OrganizationID: a.OrganizationID.String(),
	}
}

type Reader struct {
	database *gorm.DB
	ledger   repository.LedgerStore
}

func (r *Reader) Portfolio(ctx context.Context, actor Actor) (Portfolio, error) {
	var portfolio Portfolio

	err := database.WithinTenant(ctx, r.database, actor.tenancy(), func(tx database.Tx) error {
		holdings, err := r.ledger.Holdings(tx, actor.OrganizationID)
		if err != nil {
			return err
		}

		available, escrowed, retired := decimal.Zero, decimal.Zero, decimal.Zero

		for _, holding := range holdings {
			available = available.Add(decimal.RequireFromString(holding.Available))
			escrowed = escrowed.Add(decimal.RequireFromString(holding.Escrowed))
			retired = retired.Add(decimal.RequireFromString(holding.Retired))
		}

		portfolio = Portfolio{
			Holdings:       holdings,
			TotalAvailable: available.StringFixed(domain.Places),
			TotalEscrowed:  escrowed.StringFixed(domain.Places),
			TotalRetired:   retired.StringFixed(domain.Places),
		}
		return nil
	})
	if err != nil {
		return Portfolio{}, err
	}

	return portfolio, nil
}

func (r *Reader) Issuances(
	ctx context.Context,
	actor Actor,
) ([]domain.CreditIssuance, error) {
	var issuances []domain.CreditIssuance

	err := database.WithinTenant(ctx, r.database, actor.tenancy(), func(tx database.Tx) error {
		loaded, err := r.ledger.Issuances(tx, actor.OrganizationID, maximumIssuances)
		if err != nil {
			return err
		}
		issuances = loaded
		return nil
	})
	if err != nil {
		return nil, err
	}

	return issuances, nil
}

func NewReader(handle *gorm.DB, ledger repository.LedgerStore) *Reader {
	return &Reader{database: handle, ledger: ledger}
}
