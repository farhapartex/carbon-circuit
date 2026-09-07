package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/domain"
)

type Session struct {
	domain.Base
	UserID         uuid.UUID  `gorm:"column:user_id;type:uuid"`
	Auth0SessionID string     `gorm:"column:auth0_session_id"`
	UserAgent      string     `gorm:"column:user_agent"`
	IPAddress      string     `gorm:"column:ip_address;type:inet"`
	StartedAt      time.Time  `gorm:"column:started_at"`
	LastSeenAt     time.Time  `gorm:"column:last_seen_at"`
	RevokedAt      *time.Time `gorm:"column:revoked_at"`
}

func (Session) TableName() string { return "sessions" }
