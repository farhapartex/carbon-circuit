package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/domain"
)

type OrganizationRole string

const (
	RoleOwner  OrganizationRole = "owner"
	RoleAdmin  OrganizationRole = "admin"
	RoleMember OrganizationRole = "member"
)

type MembershipState string

const (
	MembershipActive  MembershipState = "active"
	MembershipRevoked MembershipState = "revoked"
)

type OrganizationMembership struct {
	domain.Base
	OrganizationID  uuid.UUID        `gorm:"column:organization_id;type:uuid"`
	UserID          uuid.UUID        `gorm:"column:user_id;type:uuid"`
	Role            OrganizationRole `gorm:"column:role"`
	State           MembershipState  `gorm:"column:state"`
	InvitedByUserID *uuid.UUID       `gorm:"column:invited_by_user_id;type:uuid"`
	JoinedAt        *time.Time       `gorm:"column:joined_at"`
	RevokedAt       *time.Time       `gorm:"column:revoked_at"`
}

func (OrganizationMembership) TableName() string { return "organization_memberships" }
