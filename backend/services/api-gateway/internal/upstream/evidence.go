package upstream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	evidencev1 "github.com/carboncircuit/backend/gen/carboncircuit/evidence/v1"
)

const uploadChunkSize = 256 * 1024

type Evidence struct {
	connection    *grpc.ClientConn
	client        evidencev1.EvidenceServiceClient
	callTimeout   time.Duration
	uploadTimeout time.Duration
}

func DialEvidence(
	address string,
	callTimeout, uploadTimeout time.Duration,
	transport credentials.TransportCredentials,
) (*Evidence, error) {
	connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(transport))
	if err != nil {
		return nil, err
	}

	return &Evidence{
		connection:    connection,
		client:        evidencev1.NewEvidenceServiceClient(connection),
		callTimeout:   callTimeout,
		uploadTimeout: uploadTimeout,
	}, nil
}

func (e *Evidence) Close() error { return e.connection.Close() }

type Upload struct {
	Purpose           evidencev1.DocumentPurpose
	FileName          string
	DeclaredMediaType string
	IdempotencyKey    string
	Content           io.Reader
}

func (e *Evidence) UploadDocument(
	ctx context.Context,
	upload Upload,
) (*evidencev1.UploadDocumentResponse, error) {
	callCtx, cancel := callContext(ctx, upload.IdempotencyKey, e.uploadTimeout)
	defer cancel()

	stream, err := e.client.UploadDocument(callCtx)
	if err != nil {
		return nil, err
	}

	err = stream.Send(&evidencev1.UploadDocumentRequest{
		Part: &evidencev1.UploadDocumentRequest_Metadata{
			Metadata: &evidencev1.UploadMetadata{
				Purpose:           upload.Purpose,
				FileName:          upload.FileName,
				DeclaredMediaType: upload.DeclaredMediaType,
				IdempotencyKey:    upload.IdempotencyKey,
			},
		},
	})
	if err != nil {
		return nil, sendFailure(stream, err)
	}

	buffer := make([]byte, uploadChunkSize)

	for {
		read, readErr := upload.Content.Read(buffer)
		if read > 0 {
			sendErr := stream.Send(&evidencev1.UploadDocumentRequest{
				Part: &evidencev1.UploadDocumentRequest_Chunk{Chunk: buffer[:read]},
			})
			if sendErr != nil {
				return nil, sendFailure(stream, sendErr)
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			stream.CloseAndRecv()
			return nil, fmt.Errorf("read upload body: %w", readErr)
		}
	}

	return stream.CloseAndRecv()
}

func sendFailure(
	stream evidencev1.EvidenceService_UploadDocumentClient,
	err error,
) error {
	if _, received := stream.CloseAndRecv(); received != nil {
		return received
	}
	return err
}

func (e *Evidence) GetDocument(
	ctx context.Context,
	documentID string,
) (*evidencev1.GetDocumentResponse, error) {
	callCtx, cancel := callContext(ctx, "", e.callTimeout)
	defer cancel()
	return e.client.GetDocument(callCtx, &evidencev1.GetDocumentRequest{DocumentId: documentID})
}

func (e *Evidence) ListDocuments(
	ctx context.Context,
	purpose evidencev1.DocumentPurpose,
	after string,
	limit int32,
) (*evidencev1.ListDocumentsResponse, error) {
	callCtx, cancel := callContext(ctx, "", e.callTimeout)
	defer cancel()
	return e.client.ListDocuments(callCtx, &evidencev1.ListDocumentsRequest{
		Purpose: purpose,
		After:   after,
		Limit:   limit,
	})
}

func (e *Evidence) CreateDownloadLink(
	ctx context.Context,
	documentID string,
) (*evidencev1.CreateDownloadLinkResponse, error) {
	callCtx, cancel := callContext(ctx, "", e.callTimeout)
	defer cancel()
	return e.client.CreateDownloadLink(callCtx,
		&evidencev1.CreateDownloadLinkRequest{DocumentId: documentID})
}

func (e *Evidence) Ping(ctx context.Context) (*evidencev1.PingResponse, error) {
	callCtx, cancel := callContext(ctx, "", e.callTimeout)
	defer cancel()
	return e.client.Ping(callCtx, &evidencev1.PingRequest{})
}
