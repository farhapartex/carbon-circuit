package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/provenance-read-service/internal/domain"
)

type PublicStore interface {
	UpsertBatch(tx database.Tx, batch *domain.PublicBatch) error
	EnsureBatch(tx database.Tx, batchID uuid.UUID) error
	UpsertCheckpoint(tx database.Tx, checkpoint *domain.PublicCheckpoint) error
	MarkSuperseded(tx database.Tx, checkpointID uuid.UUID) error
	ApplyScore(tx database.Tx, batchID uuid.UUID, total int, components database.JSONDocument) error
	FindByReference(tx database.Tx, reference string) (domain.PublicBatch, bool, error)
	ListCheckpoints(tx database.Tx, batchID uuid.UUID) ([]domain.PublicCheckpoint, error)
}

type PublicRepository struct{}

func NewPublicRepository() *PublicRepository { return &PublicRepository{} }

func (r *PublicRepository) UpsertBatch(
	tx database.Tx,
	batch *domain.PublicBatch,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	err := tx.Session().
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"public_reference",
				"product_category",
				"component_type",
				"originating_facility_name",
				"originating_facility_country",
				"produced_at",
				"projected",
				"updated_at",
			}),
		}).
		Create(batch).Error
	if err != nil {
		return fmt.Errorf("upsert public batch: %w", err)
	}

	return nil
}

func (r *PublicRepository) EnsureBatch(
	tx database.Tx,
	batchID uuid.UUID,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	stub := domain.PublicBatch{
		PublicReference: placeholderReference(batchID),
		ScoreComponents: database.JSONDocument("[]"),
	}
	stub.ID = batchID

	err := tx.Session().
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoNothing: true}).
		Create(&stub).Error
	if err != nil {
		return fmt.Errorf("ensure public batch: %w", err)
	}

	return nil
}

func placeholderReference(batchID uuid.UUID) string {
	hex := batchID.String()
	stripped := make([]byte, 0, 32)
	for _, character := range []byte(hex) {
		if character != '-' {
			stripped = append(stripped, character)
		}
	}
	return string(stripped[:22])
}

func (r *PublicRepository) UpsertCheckpoint(
	tx database.Tx,
	checkpoint *domain.PublicCheckpoint,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	err := tx.Session().
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoNothing: true}).
		Create(checkpoint).Error
	if err != nil {
		return fmt.Errorf("upsert public checkpoint: %w", err)
	}

	return nil
}

func (r *PublicRepository) MarkSuperseded(
	tx database.Tx,
	checkpointID uuid.UUID,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	err := tx.Session().Model(&domain.PublicCheckpoint{}).
		Where("id = ?", checkpointID).
		Updates(map[string]any{
			"superseded": true,
			"updated_at": gorm.Expr("now()"),
			"version":    gorm.Expr("version + 1"),
		}).Error
	if err != nil {
		return fmt.Errorf("mark public checkpoint superseded: %w", err)
	}

	return nil
}

func (r *PublicRepository) ApplyScore(
	tx database.Tx,
	batchID uuid.UUID,
	total int,
	components database.JSONDocument,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	err := tx.Session().Model(&domain.PublicBatch{}).
		Where("id = ?", batchID).
		Updates(map[string]any{
			"provenance_score": total,
			"score_components": components,
			"last_updated_at":  gorm.Expr("now()"),
			"updated_at":       gorm.Expr("now()"),
			"version":          gorm.Expr("version + 1"),
		}).Error
	if err != nil {
		return fmt.Errorf("apply score to public batch: %w", err)
	}

	return nil
}

func (r *PublicRepository) FindByReference(
	tx database.Tx,
	reference string,
) (domain.PublicBatch, bool, error) {
	if err := tx.Bound(); err != nil {
		return domain.PublicBatch{}, false, err
	}

	var batch domain.PublicBatch

	err := tx.Session().
		Where("public_reference = ? AND projected", reference).
		First(&batch).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.PublicBatch{}, false, nil
	}
	if err != nil {
		return domain.PublicBatch{}, false, fmt.Errorf("find public batch: %w", err)
	}

	return batch, true, nil
}

func (r *PublicRepository) ListCheckpoints(
	tx database.Tx,
	batchID uuid.UUID,
) ([]domain.PublicCheckpoint, error) {
	if err := tx.Bound(); err != nil {
		return nil, err
	}

	var checkpoints []domain.PublicCheckpoint
	err := tx.Session().
		Where("batch_id = ?", batchID).
		Order("occurred_at, id").
		Find(&checkpoints).Error
	if err != nil {
		return nil, fmt.Errorf("list public checkpoints: %w", err)
	}

	return checkpoints, nil
}
