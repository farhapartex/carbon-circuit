package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/evidence-service/internal/domain"
)

type Page struct {
	Documents  []domain.Document
	NextCursor string
}

func (s *DocumentService) Get(
	ctx context.Context,
	actor Actor,
	documentID uuid.UUID,
) (domain.Document, error) {
	var document domain.Document

	err := database.WithinTenant(ctx, s.database, tenancy(actor), func(tx database.Tx) error {
		found, present, err := s.documents.Find(tx, actor.OrganizationID, documentID)
		if err != nil {
			return err
		}
		if !present {
			return ErrDocumentNotFound
		}
		document = found
		return nil
	})
	if err != nil {
		return domain.Document{}, err
	}

	return document, nil
}

func (s *DocumentService) List(
	ctx context.Context,
	actor Actor,
	purpose domain.Purpose,
	after string,
	limit int,
) (Page, error) {
	if limit <= 0 {
		limit = defaultPageSize
	}
	if limit > maximumPageSize {
		limit = maximumPageSize
	}

	var page Page

	err := database.WithinTenant(ctx, s.database, tenancy(actor), func(tx database.Tx) error {
		documents, err := s.documents.List(tx, actor.OrganizationID, purpose, after, limit+1)
		if err != nil {
			return err
		}

		if len(documents) > limit {
			page.NextCursor = documents[limit-1].ID.String()
			documents = documents[:limit]
		}

		page.Documents = documents
		return nil
	})
	if err != nil {
		return Page{}, err
	}

	return page, nil
}

func (s *DocumentService) DownloadLink(
	ctx context.Context,
	actor Actor,
	documentID uuid.UUID,
) (Link, domain.Document, error) {
	document, err := s.Get(ctx, actor, documentID)
	if err != nil {
		return Link{}, domain.Document{}, err
	}

	if !document.Usable() {
		return Link{}, domain.Document{}, fmt.Errorf("%w: verdict is %s",
			ErrNotDownloadable, document.ScanVerdict)
	}

	url, expiresAt, err := s.objects.SignedLink(
		ctx, *document.StorageKey, document.FileName, document.DetectedMediaType,
	)
	if err != nil {
		return Link{}, domain.Document{}, err
	}

	return Link{URL: url, ExpiresAt: expiresAt}, document, nil
}

type Resolution struct {
	DocumentID uuid.UUID
	Usable     bool
	Refusal    string
	Document   domain.Document
}

func (s *DocumentService) Resolve(
	ctx context.Context,
	actor Actor,
	documentIDs []uuid.UUID,
	purpose domain.Purpose,
) ([]Resolution, error) {
	var found []domain.Document

	err := database.WithinTenant(ctx, s.database, tenancy(actor), func(tx database.Tx) error {
		documents, err := s.documents.FindMany(tx, actor.OrganizationID, documentIDs)
		if err != nil {
			return err
		}
		found = documents
		return nil
	})
	if err != nil {
		return nil, err
	}

	byID := make(map[uuid.UUID]domain.Document, len(found))
	for _, document := range found {
		byID[document.ID] = document
	}

	resolutions := make([]Resolution, 0, len(documentIDs))

	for _, documentID := range documentIDs {
		resolutions = append(resolutions, resolve(byID, documentID, purpose))
	}

	return resolutions, nil
}

func resolve(
	byID map[uuid.UUID]domain.Document,
	documentID uuid.UUID,
	purpose domain.Purpose,
) Resolution {
	document, present := byID[documentID]
	if !present {
		return Resolution{
			DocumentID: documentID,
			Refusal:    "no such document belongs to this organization",
		}
	}
	if !document.Usable() {
		return Resolution{
			DocumentID: documentID,
			Refusal:    fmt.Sprintf("document did not pass scanning: %s", document.ScanVerdict),
			Document:   document,
		}
	}
	if purpose != "" && document.Purpose != purpose {
		return Resolution{
			DocumentID: documentID,
			Refusal: fmt.Sprintf("document was uploaded for %s, not %s",
				document.Purpose, purpose),
			Document: document,
		}
	}

	return Resolution{DocumentID: documentID, Usable: true, Document: document}
}
