package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/domain"
)

type SubscriptionState string

const (
	SubscriptionActive      SubscriptionState = "active"
	SubscriptionGracePeriod SubscriptionState = "grace_period"
	SubscriptionReadOnly    SubscriptionState = "read_only"
	SubscriptionCancelled   SubscriptionState = "cancelled"
)

type Subscription struct {
	domain.Base
	OrganizationID     uuid.UUID         `gorm:"column:organization_id;type:uuid"`
	PlanID             uuid.UUID         `gorm:"column:plan_id;type:uuid"`
	State              SubscriptionState `gorm:"column:state"`
	CurrentPeriodStart time.Time         `gorm:"column:current_period_start"`
	CurrentPeriodEnd   time.Time         `gorm:"column:current_period_end"`
	CancelledAt        *time.Time        `gorm:"column:cancelled_at"`
	Plan               Plan              `gorm:"foreignKey:PlanID;references:ID"`
}

func (Subscription) TableName() string { return "subscriptions" }

func (s Subscription) Entitles() bool {
	return s.State != SubscriptionCancelled
}
