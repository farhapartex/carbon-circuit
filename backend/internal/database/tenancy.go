package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	userSetting         = "app.user_id"
	organizationSetting = "app.organization_id"
)

type TenantContext struct {
	UserID         string
	OrganizationID string
}

func WithinTenant(
	ctx context.Context,
	database *gorm.DB,
	tenant TenantContext,
	work func(tx *gorm.DB) error,
) error {
	settings, err := tenant.settings()
	if err != nil {
		return err
	}

	return database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for setting, value := range settings {
			if err := applyLocalSetting(tx, setting, value); err != nil {
				return err
			}
		}
		return work(tx)
	})
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

	if len(settings) == 0 {
		return nil, fmt.Errorf("tenant context requires a user or organization")
	}

	return settings, nil
}

func applyLocalSetting(tx *gorm.DB, setting, value string) error {
	if err := tx.Exec("SELECT set_config(?, ?, true)", setting, value).Error; err != nil {
		return fmt.Errorf("apply %s: %w", setting, err)
	}
	return nil
}
