package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/provenance-service/internal/domain"
)

var (
	ErrExternalIDTaken  = errors.New("external id already used by this organization")
	ErrConcurrentUpdate = errors.New("batch changed while this update was being prepared")
)

type BatchStore interface {
	Insert(tx database.Tx, batch *domain.Batch) error
	InsertParent(tx database.Tx, parent *domain.BatchParent) error
	Find(tx database.Tx, batchID uuid.UUID) (domain.Batch, bool, error)
	FindOwned(tx database.Tx, organizationID, batchID uuid.UUID) (domain.Batch, bool, error)
	List(tx database.Tx, organizationID uuid.UUID, after string, limit int) ([]domain.Batch, error)
	ResolveReference(tx database.Tx, declared string) (uuid.UUID, bool, error)
	ListParents(tx database.Tx, batchID uuid.UUID) ([]domain.BatchParent, error)
	AncestorDepth(tx database.Tx, batchID uuid.UUID) (int, error)
	UpdateScore(tx database.Tx, batch *domain.Batch) error
}

type BatchRepository struct{}

func NewBatchRepository() *BatchRepository { return &BatchRepository{} }

func (r *BatchRepository) Insert(tx database.Tx, batch *domain.Batch) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	err := tx.Session().Create(batch).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrExternalIDTaken
	}
	if err != nil {
		return fmt.Errorf("insert batch: %w", err)
	}

	return nil
}

func (r *BatchRepository) InsertParent(
	tx database.Tx,
	parent *domain.BatchParent,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	if err := tx.Session().Create(parent).Error; err != nil {
		return fmt.Errorf("insert batch parent: %w", err)
	}

	return nil
}

func (r *BatchRepository) Find(
	tx database.Tx,
	batchID uuid.UUID,
) (domain.Batch, bool, error) {
	if err := tx.Bound(); err != nil {
		return domain.Batch{}, false, err
	}

	var batch domain.Batch

	err := tx.Session().First(&batch, "id = ?", batchID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Batch{}, false, nil
	}
	if err != nil {
		return domain.Batch{}, false, fmt.Errorf("find batch: %w", err)
	}

	return batch, true, nil
}

func (r *BatchRepository) FindOwned(
	tx database.Tx,
	organizationID, batchID uuid.UUID,
) (domain.Batch, bool, error) {
	if err := tx.Bound(); err != nil {
		return domain.Batch{}, false, err
	}

	var batch domain.Batch

	err := tx.Session().
		Where("id = ? AND organization_id = ?", batchID, organizationID).
		First(&batch).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Batch{}, false, nil
	}
	if err != nil {
		return domain.Batch{}, false, fmt.Errorf("find owned batch: %w", err)
	}

	return batch, true, nil
}

func (r *BatchRepository) List(
	tx database.Tx,
	organizationID uuid.UUID,
	after string,
	limit int,
) ([]domain.Batch, error) {
	if err := tx.Bound(); err != nil {
		return nil, err
	}

	query := tx.Session().
		Where("organization_id = ?", organizationID).
		Order("id DESC").
		Limit(limit)

	if after != "" {
		query = query.Where("id < ?", after)
	}

	var batches []domain.Batch
	if err := query.Find(&batches).Error; err != nil {
		return nil, fmt.Errorf("list batches: %w", err)
	}

	return batches, nil
}

func (r *BatchRepository) ResolveReference(
	tx database.Tx,
	declared string,
) (uuid.UUID, bool, error) {
	if err := tx.Bound(); err != nil {
		return uuid.Nil, false, err
	}

	if err := tx.Session().Exec(
		"SELECT set_config('app.resolving_reference', ?, true)", declared,
	).Error; err != nil {
		return uuid.Nil, false, fmt.Errorf("scope reference resolution: %w", err)
	}

	var found []domain.Batch
	err := tx.Session().
		Select("id").
		Where("public_reference = ? AND deleted_at IS NULL", declared).
		Limit(1).
		Find(&found).Error

	if clearErr := tx.Session().Exec(
		"SELECT set_config('app.resolving_reference', '', true)",
	).Error; clearErr != nil {
		return uuid.Nil, false, fmt.Errorf("clear reference resolution: %w", clearErr)
	}

	if err != nil {
		return uuid.Nil, false, fmt.Errorf("resolve public reference: %w", err)
	}
	if len(found) == 0 {
		return uuid.Nil, false, nil
	}

	return found[0].ID, true, nil
}

func (r *BatchRepository) ListParents(
	tx database.Tx,
	batchID uuid.UUID,
) ([]domain.BatchParent, error) {
	if err := tx.Bound(); err != nil {
		return nil, err
	}

	var parents []domain.BatchParent
	err := tx.Session().
		Where("batch_id = ?", batchID).
		Order("created_at").
		Find(&parents).Error
	if err != nil {
		return nil, fmt.Errorf("list batch parents: %w", err)
	}

	return parents, nil
}

func (r *BatchRepository) AncestorDepth(
	tx database.Tx,
	batchID uuid.UUID,
) (int, error) {
	if err := tx.Bound(); err != nil {
		return 0, err
	}

	var depth int
	err := tx.Session().Raw(`
		WITH RECURSIVE ancestry(id, depth) AS (
			SELECT ?::uuid, 0
			UNION ALL
			SELECT parent.parent_batch_id, ancestry.depth + 1
			FROM provenance.batch_parents parent
			JOIN ancestry ON parent.batch_id = ancestry.id
			WHERE parent.parent_batch_id IS NOT NULL
			  AND parent.deleted_at IS NULL
			  AND ancestry.depth < ?
		)
		SELECT coalesce(max(depth), 0) FROM ancestry
	`, batchID, domain.MaximumParentChainDepth+1).Scan(&depth).Error
	if err != nil {
		return 0, fmt.Errorf("measure ancestor depth: %w", err)
	}

	return depth, nil
}

func (r *BatchRepository) UpdateScore(
	tx database.Tx,
	batch *domain.Batch,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	result := tx.Session().Model(&domain.Batch{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND version = ?", batch.ID, batch.Version).
		Updates(map[string]any{
			"checkpoint_count": batch.CheckpointCount,
			"provenance_score": batch.ProvenanceScore,
			"score_components": batch.ScoreComponents,
			"version":          batch.Version + 1,
			"updated_at":       gorm.Expr("now()"),
		})

	if result.Error != nil {
		return fmt.Errorf("update batch score: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrConcurrentUpdate
	}

	batch.Version++

	return nil
}
