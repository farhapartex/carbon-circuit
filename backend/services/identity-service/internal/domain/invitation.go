package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/domain"
)

type InvitationState string

const (
	InvitationPending  InvitationState = "pending"
	InvitationAccepted InvitationState = "accepted"
	InvitationRevoked  InvitationState = "revoked"
	InvitationExpired  InvitationState = "expired"
)

type Invitation struct {
	domain.Base
	OrganizationID   uuid.UUID        `gorm:"column:organization_id;type:uuid"`
	Email            string           `gorm:"column:email;type:citext"`
	Role             OrganizationRole `gorm:"column:role"`
	TokenHash        []byte           `gorm:"column:token_hash"`
	State            InvitationState  `gorm:"column:state"`
	InvitedByUserID  uuid.UUID        `gorm:"column:invited_by_user_id;type:uuid"`
	ExpiresAt        time.Time        `gorm:"column:expires_at"`
	AcceptedAt       *time.Time       `gorm:"column:accepted_at"`
	AcceptedByUserID *uuid.UUID       `gorm:"column:accepted_by_user_id;type:uuid"`
}

func (Invitation) TableName() string { return "invitations" }

func (i Invitation) Live(now time.Time) bool {
	return i.State == InvitationPending && now.Before(i.ExpiresAt)
}
