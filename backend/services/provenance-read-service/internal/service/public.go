package service

import (
	"context"
	"errors"
	"log/slog"

	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/provenance-read-service/internal/domain"
	"github.com/carboncircuit/backend/services/provenance-read-service/internal/repository"
)

var ErrBatchNotFound = errors.New("no batch carries that reference")

type PublicView struct {
	Batch       domain.PublicBatch
	Checkpoints []domain.PublicCheckpoint
}

type PublicReader struct {
	database *gorm.DB
	store    repository.PublicStore
	logger   *slog.Logger
}

func NewPublicReader(
	handle *gorm.DB,
	store repository.PublicStore,
	logger *slog.Logger,
) *PublicReader {
	return &PublicReader{database: handle, store: store, logger: logger}
}

func (r *PublicReader) Track(
	ctx context.Context,
	reference string,
) (PublicView, error) {
	var view PublicView

	err := database.Within(ctx, r.database, func(tx database.Tx) error {
		batch, found, err := r.store.FindByReference(tx, reference)
		if err != nil {
			return err
		}
		if !found {
			return ErrBatchNotFound
		}

		checkpoints, err := r.store.ListCheckpoints(tx, batch.ID)
		if err != nil {
			return err
		}

		view = PublicView{Batch: batch, Checkpoints: checkpoints}

		return nil
	})
	if err != nil {
		return PublicView{}, err
	}

	return view, nil
}
