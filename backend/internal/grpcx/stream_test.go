package grpcx_test

import (
	"context"
	"crypto/ed25519"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/logging"
	"github.com/carboncircuit/backend/internal/servicetoken"
)

type fakeStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s fakeStream) Context() context.Context { return s.ctx }

func signer(t *testing.T) (*servicetoken.Signer, *servicetoken.Verifier) {
	t.Helper()

	public, private, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	issuer, err := servicetoken.NewSigner(private, time.Minute)
	if err != nil {
		t.Fatalf("new signer: %v", err)
	}

	return issuer, servicetoken.NewVerifier(public)
}

func probeCaller() servicetoken.Caller {
	return servicetoken.Caller{
		Subject:        "auth0|probe",
		UserID:         "3f0f2f6c-0c4a-4f8f-9a3b-1a2b3c4d5e6f",
		OrganizationID: "8d1c2b3a-4e5f-6789-0abc-def012345678",
	}
}

func incoming(t *testing.T, token string) context.Context {
	t.Helper()

	return metadata.NewIncomingContext(context.Background(),
		metadata.Pairs(grpcx.ServiceTokenMetadataKey, token))
}

func correlationOf(ctx context.Context) string {
	return logging.CorrelationIDFrom(ctx)
}

func streamInfo() *grpc.StreamServerInfo {
	return &grpc.StreamServerInfo{
		FullMethod:     "/carboncircuit.evidence.v1.EvidenceService/UploadDocument",
		IsClientStream: true,
	}
}

func TestStreamWithoutAServiceTokenIsRefused(t *testing.T) {
	_, verifier := signer(t)

	reached := false
	handler := func(any, grpc.ServerStream) error {
		reached = true
		return nil
	}

	err := grpcx.RequireServiceTokenStream(verifier, nil)(
		nil, fakeStream{ctx: context.Background()}, streamInfo(), handler)

	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", err)
	}
	if reached {
		t.Fatal("the handler ran for a stream with no service token")
	}
}

func TestStreamWithATamperedServiceTokenIsRefused(t *testing.T) {
	issuer, verifier := signer(t)

	token, err := issuer.Issue(probeCaller())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	reached := false
	handler := func(any, grpc.ServerStream) error {
		reached = true
		return nil
	}

	tampered := token[:len(token)-4] + "AAAA"
	ctx := incoming(t, tampered)

	streamErr := grpcx.RequireServiceTokenStream(verifier, nil)(
		nil, fakeStream{ctx: ctx}, streamInfo(), handler)

	if status.Code(streamErr) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", streamErr)
	}
	if reached {
		t.Fatal("the handler ran for a stream with a tampered service token")
	}
}

func TestVerifiedStreamCarriesTheCallerForward(t *testing.T) {
	issuer, verifier := signer(t)

	token, err := issuer.Issue(probeCaller())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	var seen servicetoken.Caller
	present := false

	handler := func(_ any, stream grpc.ServerStream) error {
		seen, present = grpcx.CallerFrom(stream.Context())
		return nil
	}

	streamErr := grpcx.RequireServiceTokenStream(verifier, nil)(
		nil, fakeStream{ctx: incoming(t, token)}, streamInfo(), handler)
	if streamErr != nil {
		t.Fatalf("expected the stream to be served, got %v", streamErr)
	}

	if !present {
		t.Fatal("the verified caller did not reach the stream handler")
	}
	if seen.OrganizationID != "8d1c2b3a-4e5f-6789-0abc-def012345678" {
		t.Fatalf("caller carried organization %q", seen.OrganizationID)
	}
}

func TestExemptStreamMethodSkipsVerification(t *testing.T) {
	_, verifier := signer(t)

	reached := false
	handler := func(any, grpc.ServerStream) error {
		reached = true
		return nil
	}

	exempt := map[string]bool{streamInfo().FullMethod: true}

	err := grpcx.RequireServiceTokenStream(verifier, exempt)(
		nil, fakeStream{ctx: context.Background()}, streamInfo(), handler)
	if err != nil {
		t.Fatalf("expected an exempt method to be served, got %v", err)
	}
	if !reached {
		t.Fatal("an exempt method must reach its handler")
	}
}

func TestPanickingStreamBecomesInternalRatherThanCrashing(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	handler := func(any, grpc.ServerStream) error {
		panic("evidence stream exploded")
	}

	err := grpcx.RecoverStream(logger)(
		nil, fakeStream{ctx: context.Background()}, streamInfo(), handler)

	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal, got %v", err)
	}
}

func TestStreamHandlerErrorIsReturnedUnchanged(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	sentinel := status.Error(codes.InvalidArgument, "declared media type does not match content")

	handler := func(any, grpc.ServerStream) error { return sentinel }

	err := grpcx.LogStream(logger)(
		nil, fakeStream{ctx: context.Background()}, streamInfo(), handler)

	if !errors.Is(err, sentinel) && err.Error() != sentinel.Error() {
		t.Fatalf("expected the handler error to survive logging, got %v", err)
	}
}

func TestCorrelateStreamAlwaysProvidesAnIdentifier(t *testing.T) {
	seen := ""

	handler := func(_ any, stream grpc.ServerStream) error {
		seen = correlationOf(stream.Context())
		return nil
	}

	if err := grpcx.CorrelateStream()(
		nil, fakeStream{ctx: context.Background()}, streamInfo(), handler); err != nil {
		t.Fatalf("correlate: %v", err)
	}

	if seen == "" {
		t.Fatal("a stream with no incoming request id must still be correlated")
	}
}
