package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

type UserStore interface {
	FindByAuth0Subject(ctx context.Context, subject string) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	AttachAuth0Subject(ctx context.Context, userID uuid.UUID, subject string) error
}

type UserRepository struct {
	database *gorm.DB
}

func NewUserRepository(database *gorm.DB) *UserRepository {
	return &UserRepository{database: database}
}

func (r *UserRepository) FindByAuth0Subject(ctx context.Context, subject string) (domain.User, error) {
	return r.findBy(ctx, "auth0_subject = ?", subject)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	return r.findBy(ctx, "email = ?", email)
}

func (r *UserRepository) findBy(ctx context.Context, condition string, argument any) (domain.User, error) {
	var user domain.User

	err := r.database.WithContext(ctx).Where(condition, argument).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.User{}, ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("find user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	err := r.database.WithContext(ctx).Create(user).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrUserExists
	}
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) AttachAuth0Subject(
	ctx context.Context,
	userID uuid.UUID,
	subject string,
) error {
	result := r.database.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ? AND auth0_subject IS NULL", userID).
		Updates(map[string]any{"auth0_subject": subject, "email_verified": true})

	if result.Error != nil {
		return fmt.Errorf("attach auth0 subject: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("attach auth0 subject: user %s already carries a subject", userID)
	}

	return nil
}
