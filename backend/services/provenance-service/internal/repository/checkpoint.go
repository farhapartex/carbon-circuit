package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/provenance-service/internal/domain"
)

type CheckpointStore interface {
	Insert(tx database.Tx, checkpoint *domain.Checkpoint) error
	ListForBatch(tx database.Tx, batchID uuid.UUID) ([]domain.Checkpoint, error)
	PageForBatch(
		tx database.Tx,
		batchID uuid.UUID,
		after string,
		limit int,
	) ([]domain.Checkpoint, error)
	Find(tx database.Tx, checkpointID uuid.UUID) (domain.Checkpoint, bool, error)
	MarkSuperseded(tx database.Tx, originalID, correctionID uuid.UUID) error
}

type CheckpointRepository struct{}

func NewCheckpointRepository() *CheckpointRepository { return &CheckpointRepository{} }

func (r *CheckpointRepository) Insert(
	tx database.Tx,
	checkpoint *domain.Checkpoint,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	if err := tx.Session().Create(checkpoint).Error; err != nil {
		return fmt.Errorf("insert checkpoint: %w", err)
	}

	return nil
}

func (r *CheckpointRepository) ListForBatch(
	tx database.Tx,
	batchID uuid.UUID,
) ([]domain.Checkpoint, error) {
	if err := tx.Bound(); err != nil {
		return nil, err
	}

	var checkpoints []domain.Checkpoint
	err := tx.Session().
		Where("batch_id = ?", batchID).
		Order("occurred_at, id").
		Find(&checkpoints).Error
	if err != nil {
		return nil, fmt.Errorf("list checkpoints: %w", err)
	}

	return checkpoints, nil
}

func (r *CheckpointRepository) PageForBatch(
	tx database.Tx,
	batchID uuid.UUID,
	after string,
	limit int,
) ([]domain.Checkpoint, error) {
	if err := tx.Bound(); err != nil {
		return nil, err
	}

	query := tx.Session().
		Where("batch_id = ?", batchID).
		Order("id").
		Limit(limit)

	if after != "" {
		query = query.Where("id > ?", after)
	}

	var checkpoints []domain.Checkpoint
	if err := query.Find(&checkpoints).Error; err != nil {
		return nil, fmt.Errorf("page checkpoints: %w", err)
	}

	return checkpoints, nil
}

func (r *CheckpointRepository) Find(
	tx database.Tx,
	checkpointID uuid.UUID,
) (domain.Checkpoint, bool, error) {
	if err := tx.Bound(); err != nil {
		return domain.Checkpoint{}, false, err
	}

	var checkpoint domain.Checkpoint

	err := tx.Session().First(&checkpoint, "id = ?", checkpointID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Checkpoint{}, false, nil
	}
	if err != nil {
		return domain.Checkpoint{}, false, fmt.Errorf("find checkpoint: %w", err)
	}

	return checkpoint, true, nil
}

func (r *CheckpointRepository) MarkSuperseded(
	tx database.Tx,
	originalID, correctionID uuid.UUID,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	result := tx.Session().Model(&domain.Checkpoint{}).
		Where("id = ? AND superseded_by_checkpoint_id IS NULL", originalID).
		Updates(map[string]any{
			"superseded_by_checkpoint_id": correctionID,
			"updated_at":                  gorm.Expr("now()"),
			"version":                     gorm.Expr("version + 1"),
		})

	if result.Error != nil {
		return fmt.Errorf("mark checkpoint superseded: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrAlreadyCorrected
	}

	return nil
}

var ErrAlreadyCorrected = errors.New("checkpoint has already been corrected")
