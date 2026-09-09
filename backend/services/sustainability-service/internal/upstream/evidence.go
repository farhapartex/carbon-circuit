package upstream

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	evidencev1 "github.com/carboncircuit/backend/gen/carboncircuit/evidence/v1"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/service"
)

type Evidence struct {
	connection  *grpc.ClientConn
	client      evidencev1.EvidenceServiceClient
	callTimeout time.Duration
}

func DialEvidence(
	address string,
	callTimeout time.Duration,
	transport credentials.TransportCredentials,
) (*Evidence, error) {
	connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(transport))
	if err != nil {
		return nil, err
	}

	return &Evidence{
		connection:  connection,
		client:      evidencev1.NewEvidenceServiceClient(connection),
		callTimeout: callTimeout,
	}, nil
}

func (e *Evidence) Close() error { return e.connection.Close() }

func (e *Evidence) Resolve(
	ctx context.Context,
	documentIDs []uuid.UUID,
) ([]service.ResolvedEvidence, error) {
	callCtx, cancel := context.WithTimeout(
		grpcx.ForwardServiceToken(ctx), e.callTimeout,
	)
	defer cancel()

	requested := make([]string, 0, len(documentIDs))
	for _, documentID := range documentIDs {
		requested = append(requested, documentID.String())
	}

	response, err := e.client.ResolveDocuments(callCtx, &evidencev1.ResolveDocumentsRequest{
		DocumentIds: requested,
		Purpose:     evidencev1.DocumentPurpose_DOCUMENT_PURPOSE_CLAIM_EVIDENCE,
	})
	if err != nil {
		return nil, fmt.Errorf("resolve evidence: %w", err)
	}

	resolved := make([]service.ResolvedEvidence, 0, len(response.GetDocuments()))

	for _, document := range response.GetDocuments() {
		documentID, parseErr := uuid.Parse(document.GetDocumentId())
		if parseErr != nil {
			return nil, fmt.Errorf("evidence returned an unusable document id")
		}

		var pages *int
		if document.GetPageCount() > 0 {
			counted := int(document.GetPageCount())
			pages = &counted
		}

		resolved = append(resolved, service.ResolvedEvidence{
			DocumentID:  documentID,
			Usable:      document.GetUsable(),
			Refusal:     document.GetRefusal(),
			ContentHash: document.GetContentHash(),
			FileName:    document.GetFileName(),
			MediaType:   document.GetMediaType(),
			ByteSize:    document.GetByteSize(),
			PageCount:   pages,
		})
	}

	return resolved, nil
}

func (e *Evidence) Ping(ctx context.Context) error {
	callCtx, cancel := context.WithTimeout(ctx, e.callTimeout)
	defer cancel()

	_, err := e.client.Ping(callCtx, &evidencev1.PingRequest{})
	return err
}
