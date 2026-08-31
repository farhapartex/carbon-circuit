package auth

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"

	"github.com/carboncircuit/backend/internal/cache"
)

func newDenylist(t *testing.T) (*Denylist, *miniredis.Miniredis) {
	t.Helper()

	server := miniredis.RunT(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := cache.New(server.Addr(), "", 0, logger)

	return NewDenylist(client, logger, 15*time.Minute), server
}

func TestUnknownSubjectIsAdmitted(t *testing.T) {
	denylist, _ := newDenylist(t)

	caller := Caller{Subject: testSubject, IssuedAt: time.Now()}
	if denylist.Revoked(context.Background(), caller) {
		t.Fatal("expected a subject with no revocation entry to be admitted")
	}
}

func TestTokenIssuedBeforeRevocationIsRejected(t *testing.T) {
	denylist, _ := newDenylist(t)
	ctx := context.Background()

	caller := Caller{Subject: testSubject, IssuedAt: time.Now().Add(-time.Minute)}
	if err := denylist.Revoke(ctx, testSubject); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	if !denylist.Revoked(ctx, caller) {
		t.Fatal("expected a token issued before revocation to be rejected")
	}
}

func TestTokenIssuedAfterRevocationIsAdmitted(t *testing.T) {
	denylist, _ := newDenylist(t)
	ctx := context.Background()

	if err := denylist.Revoke(ctx, testSubject); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	caller := Caller{Subject: testSubject, IssuedAt: time.Now().Add(time.Minute)}
	if denylist.Revoked(ctx, caller) {
		t.Fatal("expected a token issued after revocation to be admitted")
	}
}

func TestRevocationOfOneSubjectLeavesOthersAdmitted(t *testing.T) {
	denylist, _ := newDenylist(t)
	ctx := context.Background()

	if err := denylist.Revoke(ctx, testSubject); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	other := Caller{Subject: "auth0|someone-else", IssuedAt: time.Now().Add(-time.Minute)}
	if denylist.Revoked(ctx, other) {
		t.Fatal("expected an unrelated subject to remain admitted")
	}
}

func TestDenylistAdmitsWhenRedisIsUnreachable(t *testing.T) {
	denylist, server := newDenylist(t)
	ctx := context.Background()

	if err := denylist.Revoke(ctx, testSubject); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	server.Close()

	caller := Caller{Subject: testSubject, IssuedAt: time.Now().Add(-time.Minute)}
	if denylist.Revoked(ctx, caller) {
		t.Fatal("expected the denylist to fail open when redis is unreachable")
	}
}
