package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/provenance-service/internal/domain"
)

type BatchPage struct {
	Batches []domain.Batch
	Cursor  string
	HasMore bool
}

func pageSize(requested int) int {
	if requested <= 0 {
		return defaultPageSize
	}
	if requested > maximumPageSize {
		return maximumPageSize
	}
	return requested
}

func (s *BatchService) List(
	ctx context.Context,
	actor Actor,
	after string,
	limit int,
) (BatchPage, error) {
	size := pageSize(limit)

	var page BatchPage

	err := s.tenant(ctx, actor, func(tx database.Tx) error {
		batches, err := s.batches.List(tx, actor.OrganizationID, after, size+1)
		if err != nil {
			return err
		}

		page.HasMore = len(batches) > size
		if page.HasMore {
			batches = batches[:size]
		}
		if len(batches) > 0 {
			page.Cursor = batches[len(batches)-1].ID.String()
		}
		page.Batches = batches

		return nil
	})

	return page, err
}

func (s *BatchService) Get(
	ctx context.Context,
	actor Actor,
	batchID uuid.UUID,
) (BatchView, error) {
	var view BatchView

	err := s.tenant(ctx, actor, func(tx database.Tx) error {
		batch, found, err := s.batches.FindOwned(tx, actor.OrganizationID, batchID)
		if err != nil {
			return err
		}
		if !found {
			return ErrBatchNotFound
		}

		declarations, err := s.batches.ListParents(tx, batchID)
		if err != nil {
			return err
		}

		parents := make([]ParentView, 0, len(declarations))
		for _, declaration := range declarations {
			parent := ParentView{DeclaredReference: declaration.DeclaredReference}

			if declaration.ParentBatchID != nil {
				disclosable, exists, err := s.batches.Find(tx, *declaration.ParentBatchID)
				if err != nil {
					return err
				}
				if exists {
					parent.Resolved = true
					parent.Batch = &disclosable
				}
			}

			parents = append(parents, parent)
		}

		view = BatchView{Batch: batch, Parents: parents}

		return nil
	})
	if err != nil {
		return BatchView{}, err
	}

	return view, nil
}

type CheckpointPage struct {
	Checkpoints []domain.Checkpoint
	Cursor      string
	HasMore     bool
}

func (s *BatchService) Checkpoints(
	ctx context.Context,
	actor Actor,
	batchID uuid.UUID,
	after string,
	limit int,
) (CheckpointPage, error) {
	size := pageSize(limit)

	var page CheckpointPage

	err := s.tenant(ctx, actor, func(tx database.Tx) error {
		if _, found, err := s.batches.Find(tx, batchID); err != nil {
			return err
		} else if !found {
			return ErrBatchNotFound
		}

		checkpoints, err := s.checkpoints.PageForBatch(tx, batchID, after, size+1)
		if err != nil {
			return err
		}

		page.HasMore = len(checkpoints) > size
		if page.HasMore {
			checkpoints = checkpoints[:size]
		}
		if len(checkpoints) > 0 {
			page.Cursor = checkpoints[len(checkpoints)-1].ID.String()
		}
		page.Checkpoints = checkpoints

		return nil
	})

	return page, err
}

type ComponentView struct {
	Batch       domain.Batch
	Checkpoints []domain.Checkpoint
}

func (s *BatchService) Component(
	ctx context.Context,
	actor Actor,
	batchID, componentBatchID uuid.UUID,
) (ComponentView, error) {
	var view ComponentView

	err := s.tenant(ctx, actor, func(tx database.Tx) error {
		if _, found, err := s.batches.FindOwned(tx, actor.OrganizationID, batchID); err != nil {
			return err
		} else if !found {
			return ErrBatchNotFound
		}

		declarations, err := s.batches.ListParents(tx, batchID)
		if err != nil {
			return err
		}

		declared := false
		for _, declaration := range declarations {
			if declaration.ParentBatchID != nil && *declaration.ParentBatchID == componentBatchID {
				declared = true
				break
			}
		}
		if !declared {
			return ErrBatchNotFound
		}

		component, found, err := s.batches.Find(tx, componentBatchID)
		if err != nil {
			return err
		}
		if !found {
			return ErrBatchNotFound
		}

		checkpoints, err := s.checkpoints.ListForBatch(tx, componentBatchID)
		if err != nil {
			return err
		}

		view = ComponentView{Batch: component, Checkpoints: checkpoints}

		return nil
	})
	if err != nil {
		return ComponentView{}, err
	}

	return view, nil
}

func (s *BatchService) tenant(
	ctx context.Context,
	actor Actor,
	work func(tx database.Tx) error,
) error {
	return database.WithinTenant(
		ctx,
		s.database,
		database.TenantContext{
			UserID:         actor.UserID.String(),
			OrganizationID: actor.OrganizationID.String(),
		},
		work,
	)
}
