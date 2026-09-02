package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/domain"
)

type SiweNonce struct {
	domain.Base
	Nonce      string     `gorm:"column:nonce"`
	Domain     string     `gorm:"column:domain"`
	UserID     *uuid.UUID `gorm:"column:user_id;type:uuid"`
	IssuedAt   time.Time  `gorm:"column:issued_at"`
	ExpiresAt  time.Time  `gorm:"column:expires_at"`
	ConsumedAt *time.Time `gorm:"column:consumed_at"`
}

func (SiweNonce) TableName() string { return "siwe_nonces" }
