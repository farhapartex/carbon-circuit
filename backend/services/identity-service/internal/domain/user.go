package domain

import (
	"time"

	"github.com/carboncircuit/backend/internal/domain"
)

type PlatformRole string

const (
	PlatformVerifier PlatformRole = "verifier"
	PlatformAdmin    PlatformRole = "platform_admin"
)

type User struct {
	domain.Base
	Auth0Subject  *string       `gorm:"column:auth0_subject"`
	Email         string        `gorm:"column:email;type:citext"`
	EmailVerified bool          `gorm:"column:email_verified"`
	Name          string        `gorm:"column:name"`
	PlatformRole  *PlatformRole `gorm:"column:platform_role"`
	MFAEnrolledAt *time.Time    `gorm:"column:mfa_enrolled_at"`
	LastActiveAt  *time.Time    `gorm:"column:last_active_at"`
}

func (User) TableName() string { return "users" }

func (u User) MFAEnrolled() bool { return u.MFAEnrolledAt != nil }
