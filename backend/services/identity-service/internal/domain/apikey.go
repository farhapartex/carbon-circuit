package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/domain"
)

type APIKey struct {
	domain.Base
	OrganizationID  uuid.UUID  `gorm:"column:organization_id;type:uuid"`
	Name            string     `gorm:"column:name"`
	Prefix          string     `gorm:"column:prefix;type:char(8)"`
	SecretHMAC      []byte     `gorm:"column:secret_hmac"`
	CreatedByUserID uuid.UUID  `gorm:"column:created_by_user_id;type:uuid"`
	LastUsedAt      *time.Time `gorm:"column:last_used_at"`
	RevokedAt       *time.Time `gorm:"column:revoked_at"`
	RevokedByUserID *uuid.UUID `gorm:"column:revoked_by_user_id;type:uuid"`
}

func (APIKey) TableName() string { return "api_keys" }

func (k APIKey) Active() bool { return k.RevokedAt == nil }
