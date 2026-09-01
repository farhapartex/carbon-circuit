package idempotency

import (
	"crypto/sha256"
	"time"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/database"
)

type State string

const (
	StateProcessing State = "processing"
	StateCompleted  State = "completed"
	StateFailed     State = "failed"
)

type Record struct {
	ID             uuid.UUID             `gorm:"column:id;type:uuid;primaryKey;default:uuidv7()"`
	CreatedAt      time.Time             `gorm:"column:created_at"`
	UpdatedAt      time.Time             `gorm:"column:updated_at"`
	Version        int                   `gorm:"column:version;default:1"`
	OrganizationID *uuid.UUID            `gorm:"column:organization_id;type:uuid"`
	UserID         *uuid.UUID            `gorm:"column:user_id;type:uuid"`
	Endpoint       string                `gorm:"column:endpoint"`
	IdempotencyKey string                `gorm:"column:idempotency_key"`
	RequestHash    []byte                `gorm:"column:request_hash"`
	State          State                 `gorm:"column:state"`
	ResponseStatus *int                  `gorm:"column:response_status"`
	ResponseBody   database.JSONDocument `gorm:"column:response_body;type:json"`
	ResourceID     *uuid.UUID            `gorm:"column:resource_id;type:uuid"`
	CompletedAt    *time.Time            `gorm:"column:completed_at"`
}

func (Record) TableName() string { return "idempotency_records" }

type Scope struct {
	OrganizationID *uuid.UUID
	UserID         *uuid.UUID
}

func ForOrganization(organizationID uuid.UUID) Scope {
	return Scope{OrganizationID: &organizationID}
}

func ForUser(userID uuid.UUID) Scope {
	return Scope{UserID: &userID}
}

type Request struct {
	Scope    Scope
	Endpoint string
	Key      string
	Body     []byte
}

func (r Request) hash() []byte {
	digest := sha256.Sum256(r.Body)
	return digest[:]
}

type Response struct {
	Status     int
	Body       []byte
	ResourceID *uuid.UUID
}
