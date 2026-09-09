package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	userSetting         = "app.user_id"
	organizationSetting = "app.organization_id"
	platformRoleSetting = "app.platform_role"
)

var ErrNotInTransaction = errors.New("no transaction bound to this handle")

type Tx struct {
	session *gorm.DB
}

func (t Tx) Session() *gorm.DB { return t.session }

func (t Tx) Bound() error {
	if t.session == nil {
		return ErrNotInTransaction
	}
	return nil
}

type TenantContext struct {
	UserID         string
	OrganizationID string
	PlatformRole   string
}

func WithinTenant(
	ctx context.Context,
	database *gorm.DB,
	tenant TenantContext,
	work func(tx Tx) error,
) error {
	settings, err := tenant.settings()
	if err != nil {
		return err
	}

	return database.WithContext(ctx).Transaction(func(session *gorm.DB) error {
		for setting, value := range settings {
			if err := applyLocalSetting(session, setting, value); err != nil {
				return err
			}
		}
		return work(Tx{session: session})
	})
}

func SettingsFor(tenant TenantContext) (map[string]string, error) {
	return tenant.settings()
}

func (t TenantContext) settings() (map[string]string, error) {
	settings := make(map[string]string, 2)

	for setting, value := range map[string]string{
		userSetting:         t.UserID,
		organizationSetting: t.OrganizationID,
	} {
		if value == "" {
			continue
		}
		if _, err := uuid.Parse(value); err != nil {
			return nil, fmt.Errorf("%s must be a uuid: %w", setting, err)
		}
		settings[setting] = value
	}

	if t.PlatformRole != "" {
		if t.UserID == "" {
			return nil, fmt.Errorf("a platform role requires the caller's user id")
		}
		settings[platformRoleSetting] = t.PlatformRole
	}

	if len(settings) == 0 {
		return nil, fmt.Errorf("tenant context requires a user or organization")
	}

	return settings, nil
}

func applyLocalSetting(session *gorm.DB, setting, value string) error {
	if err := session.Exec("SELECT set_config(?, ?, true)", setting, value).Error; err != nil {
		return fmt.Errorf("apply %s: %w", setting, err)
	}
	return nil
}

func Within(
	ctx context.Context,
	database *gorm.DB,
	work func(tx Tx) error,
) error {
	return database.WithContext(ctx).Transaction(func(session *gorm.DB) error {
		return work(Tx{session: session})
	})
}

func AdoptTenant(tx Tx, tenant TenantContext) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	settings, err := tenant.settings()
	if err != nil {
		return err
	}

	for setting, value := range settings {
		if err := applyLocalSetting(tx.session, setting, value); err != nil {
			return err
		}
	}

	return nil
}
