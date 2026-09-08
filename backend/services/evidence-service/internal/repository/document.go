package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/evidence-service/internal/domain"
)

type DocumentStore interface {
	Insert(tx database.Tx, document *domain.Document) error
	Find(tx database.Tx, organizationID, documentID uuid.UUID) (domain.Document, bool, error)
	List(tx database.Tx, organizationID uuid.UUID, purpose domain.Purpose, after string, limit int) ([]domain.Document, error)
	FindMany(tx database.Tx, organizationID uuid.UUID, documentIDs []uuid.UUID) ([]domain.Document, error)
	CountMatchingHash(tx database.Tx, contentHash string) (int64, error)
}

type DocumentRepository struct{}

func NewDocumentRepository() *DocumentRepository { return &DocumentRepository{} }

func (r *DocumentRepository) Insert(tx database.Tx, document *domain.Document) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	if err := tx.Session().Create(document).Error; err != nil {
		return fmt.Errorf("insert document: %w", err)
	}

	return nil
}

func (r *DocumentRepository) Find(
	tx database.Tx,
	organizationID, documentID uuid.UUID,
) (domain.Document, bool, error) {
	if err := tx.Bound(); err != nil {
		return domain.Document{}, false, err
	}

	var document domain.Document

	err := tx.Session().
		First(&document, "id = ? AND organization_id = ?", documentID, organizationID).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Document{}, false, nil
	}
	if err != nil {
		return domain.Document{}, false, fmt.Errorf("find document: %w", err)
	}

	return document, true, nil
}

func (r *DocumentRepository) List(
	tx database.Tx,
	organizationID uuid.UUID,
	purpose domain.Purpose,
	after string,
	limit int,
) ([]domain.Document, error) {
	if err := tx.Bound(); err != nil {
		return nil, err
	}

	query := tx.Session().
		Where("organization_id = ?", organizationID).
		Order("created_at DESC, id DESC").
		Limit(limit)

	if purpose != "" {
		query = query.Where("purpose = ?", purpose)
	}
	if after != "" {
		query = query.Where("id < ?", after)
	}

	var documents []domain.Document
	if err := query.Find(&documents).Error; err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}

	return documents, nil
}

func (r *DocumentRepository) FindMany(
	tx database.Tx,
	organizationID uuid.UUID,
	documentIDs []uuid.UUID,
) ([]domain.Document, error) {
	if err := tx.Bound(); err != nil {
		return nil, err
	}
	if len(documentIDs) == 0 {
		return nil, nil
	}

	var documents []domain.Document

	err := tx.Session().
		Where("organization_id = ? AND id IN ?", organizationID, documentIDs).
		Find(&documents).Error
	if err != nil {
		return nil, fmt.Errorf("find documents: %w", err)
	}

	return documents, nil
}

func (r *DocumentRepository) CountMatchingHash(tx database.Tx, contentHash string) (int64, error) {
	if err := tx.Bound(); err != nil {
		return 0, err
	}

	var matches int64

	err := tx.Session().
		Model(&domain.Document{}).
		Where("content_hash = ? AND scan_status = ?", contentHash, domain.ScanClean).
		Count(&matches).Error
	if err != nil {
		return 0, fmt.Errorf("count documents matching hash: %w", err)
	}

	return matches, nil
}
