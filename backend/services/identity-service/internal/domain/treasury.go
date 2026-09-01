package domain

import (
	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/domain"
)

type TreasuryState string

const (
	TreasuryActive     TreasuryState = "active"
	TreasurySuperseded TreasuryState = "superseded"
)

type TreasuryAddress struct {
	domain.Base
	OrganizationID     uuid.UUID     `gorm:"column:organization_id;type:uuid"`
	Address            string        `gorm:"column:address;type:char(42)"`
	State              TreasuryState `gorm:"column:state"`
	DesignatedByUserID uuid.UUID     `gorm:"column:designated_by_user_id;type:uuid"`
}

func (TreasuryAddress) TableName() string { return "treasury_addresses" }
