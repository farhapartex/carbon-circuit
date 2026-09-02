package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
)

var (
	ErrInvitationPending = errors.New("an invitation is already pending for that email")
	ErrInvitationUnknown = errors.New("invitation not found")
	ErrMemberUnknown     = errors.New("member not found")
	ErrAlreadyMember     = errors.New("that email already belongs to this organization")
)

type InvitationRecord struct {
	ID            uuid.UUID
	Email         string
	Role          domain.OrganizationRole
	State         domain.InvitationState
	InvitedByName string
	CreatedAt     time.Time
	ExpiresAt     time.Time
}

type MemberRecord struct {
	UserID       uuid.UUID
	Email        string
	Name         string
	Role         domain.OrganizationRole
	MFAEnrolled  bool
	JoinedAt     *time.Time
	LastActiveAt *time.Time
}

type TeamStore interface {
	ListMembers(tx database.Tx, organizationID uuid.UUID) ([]MemberRecord, error)
	ListInvitations(tx database.Tx, organizationID uuid.UUID) ([]InvitationRecord, error)
	CountActiveOwners(tx database.Tx, organizationID uuid.UUID) (int, error)
	FindMembership(tx database.Tx, organizationID, userID uuid.UUID) (domain.OrganizationMembership, error)
	CreateInvitation(tx database.Tx, invitation *domain.Invitation) error
	FindInvitationByHash(tx database.Tx, hash []byte) (domain.Invitation, error)
	MarkInvitation(tx database.Tx, invitationID uuid.UUID, state domain.InvitationState, acceptedBy *uuid.UUID) error
	UpdateRole(tx database.Tx, membershipID uuid.UUID, role domain.OrganizationRole) error
	RevokeMembership(tx database.Tx, membershipID uuid.UUID, at time.Time) error
	CreateMembership(tx database.Tx, membership *domain.OrganizationMembership) error
	FindSubject(tx database.Tx, userID uuid.UUID) (string, error)
}

type TeamRepository struct{}

func NewTeamRepository() *TeamRepository { return &TeamRepository{} }

func (r *TeamRepository) ListMembers(
	tx database.Tx,
	organizationID uuid.UUID,
) ([]MemberRecord, error) {
	if err := tx.Bound(); err != nil {
		return nil, err
	}

	var records []MemberRecord

	err := tx.Session().
		Table("identity.organization_memberships AS m").
		Select(`m.user_id, u.email, u.name, m.role,
		        u.mfa_enrolled_at IS NOT NULL AS mfa_enrolled,
		        m.joined_at, u.last_active_at`).
		Joins("JOIN identity.users u ON u.id = m.user_id").
		Where("m.organization_id = ? AND m.state = ? AND m.deleted_at IS NULL",
			organizationID, domain.MembershipActive).
		Order("m.joined_at ASC").
		Scan(&records).Error
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}

	return records, nil
}

func (r *TeamRepository) ListInvitations(
	tx database.Tx,
	organizationID uuid.UUID,
) ([]InvitationRecord, error) {
	if err := tx.Bound(); err != nil {
		return nil, err
	}

	var invitations []InvitationRecord

	err := tx.Session().
		Table("identity.invitations AS i").
		Select("i.id, i.email, i.role, i.state, u.name AS invited_by_name, i.created_at, i.expires_at").
		Joins("JOIN identity.users u ON u.id = i.invited_by_user_id").
		Where("i.organization_id = ? AND i.state = ? AND i.deleted_at IS NULL",
			organizationID, domain.InvitationPending).
		Order("i.created_at DESC").
		Scan(&invitations).Error
	if err != nil {
		return nil, fmt.Errorf("list invitations: %w", err)
	}

	return invitations, nil
}

func (r *TeamRepository) CountActiveOwners(
	tx database.Tx,
	organizationID uuid.UUID,
) (int, error) {
	if err := tx.Bound(); err != nil {
		return 0, err
	}

	var owners int64
	err := tx.Session().Model(&domain.OrganizationMembership{}).
		Where("organization_id = ? AND role = ? AND state = ?",
			organizationID, domain.RoleOwner, domain.MembershipActive).
		Count(&owners).Error
	if err != nil {
		return 0, fmt.Errorf("count owners: %w", err)
	}

	return int(owners), nil
}

func (r *TeamRepository) FindMembership(
	tx database.Tx,
	organizationID, userID uuid.UUID,
) (domain.OrganizationMembership, error) {
	if err := tx.Bound(); err != nil {
		return domain.OrganizationMembership{}, err
	}

	var membership domain.OrganizationMembership

	err := tx.Session().
		Where("organization_id = ? AND user_id = ? AND state = ?",
			organizationID, userID, domain.MembershipActive).
		First(&membership).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.OrganizationMembership{}, ErrMemberUnknown
	}
	if err != nil {
		return domain.OrganizationMembership{}, fmt.Errorf("find membership: %w", err)
	}

	return membership, nil
}

func (r *TeamRepository) CreateInvitation(
	tx database.Tx,
	invitation *domain.Invitation,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	err := tx.Session().Create(invitation).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrInvitationPending
	}
	if err != nil {
		return fmt.Errorf("create invitation: %w", err)
	}

	return nil
}

func (r *TeamRepository) FindInvitationByHash(
	tx database.Tx,
	hash []byte,
) (domain.Invitation, error) {
	if err := tx.Bound(); err != nil {
		return domain.Invitation{}, err
	}

	var invitation domain.Invitation

	err := tx.Session().First(&invitation, "token_hash = ?", hash).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Invitation{}, ErrInvitationUnknown
	}
	if err != nil {
		return domain.Invitation{}, fmt.Errorf("find invitation: %w", err)
	}

	return invitation, nil
}

func (r *TeamRepository) MarkInvitation(
	tx database.Tx,
	invitationID uuid.UUID,
	state domain.InvitationState,
	acceptedBy *uuid.UUID,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	changes := map[string]any{"state": state, "updated_at": time.Now()}
	if state == domain.InvitationAccepted {
		now := time.Now()
		changes["accepted_at"] = now
		changes["accepted_by_user_id"] = acceptedBy
	}

	result := tx.Session().Model(&domain.Invitation{}).
		Where("id = ? AND state = ?", invitationID, domain.InvitationPending).
		Updates(changes)

	if result.Error != nil {
		return fmt.Errorf("mark invitation: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrInvitationUnknown
	}

	return nil
}

func (r *TeamRepository) UpdateRole(
	tx database.Tx,
	membershipID uuid.UUID,
	role domain.OrganizationRole,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	result := tx.Session().Model(&domain.OrganizationMembership{}).
		Where("id = ? AND state = ?", membershipID, domain.MembershipActive).
		Updates(map[string]any{"role": role, "updated_at": time.Now()})

	if result.Error != nil {
		return fmt.Errorf("update role: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrMemberUnknown
	}

	return nil
}

func (r *TeamRepository) RevokeMembership(
	tx database.Tx,
	membershipID uuid.UUID,
	at time.Time,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	result := tx.Session().Model(&domain.OrganizationMembership{}).
		Where("id = ? AND state = ?", membershipID, domain.MembershipActive).
		Updates(map[string]any{
			"state":      domain.MembershipRevoked,
			"revoked_at": at,
			"updated_at": at,
		})

	if result.Error != nil {
		return fmt.Errorf("revoke membership: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrMemberUnknown
	}

	return nil
}

func (r *TeamRepository) CreateMembership(
	tx database.Tx,
	membership *domain.OrganizationMembership,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	err := tx.Session().Create(membership).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrAlreadyMember
	}
	if err != nil {
		return fmt.Errorf("create membership: %w", err)
	}

	return nil
}

func (r *TeamRepository) FindSubject(tx database.Tx, userID uuid.UUID) (string, error) {
	if err := tx.Bound(); err != nil {
		return "", err
	}

	var subject *string
	err := tx.Session().Table("identity.users").
		Select("auth0_subject").
		Where("id = ?", userID).
		Scan(&subject).Error
	if err != nil {
		return "", fmt.Errorf("find auth0 subject: %w", err)
	}

	if subject == nil {
		return "", nil
	}

	return *subject, nil
}
