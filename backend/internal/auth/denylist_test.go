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

func TestRevokingOneSessionLeavesTheOtherDeviceSignedIn(t *testing.T) {
	denylist, _ := newDenylist(t)
	issued := time.Now().Add(-time.Minute)

	laptop := Caller{Subject: testSubject, SessionID: "sess-laptop", IssuedAt: issued}
	phone := Caller{Subject: testSubject, SessionID: "sess-phone", IssuedAt: issued}

	if err := denylist.RevokeSession(context.Background(), testSubject, "sess-laptop"); err != nil {
		t.Fatalf("revoke session: %v", err)
	}

	if !denylist.Revoked(context.Background(), laptop) {
		t.Fatal("the revoked device must be refused")
	}
	if denylist.Revoked(context.Background(), phone) {
		t.Fatal("revoking one device must not sign the other one out")
	}
}

func TestSubjectRevocationStillSignsOutEveryDevice(t *testing.T) {
	denylist, _ := newDenylist(t)
	issued := time.Now().Add(-time.Minute)

	laptop := Caller{Subject: testSubject, SessionID: "sess-laptop", IssuedAt: issued}
	phone := Caller{Subject: testSubject, SessionID: "sess-phone", IssuedAt: issued}

	if err := denylist.Revoke(context.Background(), testSubject); err != nil {
		t.Fatalf("revoke subject: %v", err)
	}

	if !denylist.Revoked(context.Background(), laptop) {
		t.Fatal("a subject-wide revocation must refuse the laptop")
	}
	if !denylist.Revoked(context.Background(), phone) {
		t.Fatal("a subject-wide revocation must refuse the phone too")
	}
}

func TestSessionRevocationDoesNotReachAnotherSubject(t *testing.T) {
	denylist, _ := newDenylist(t)
	issued := time.Now().Add(-time.Minute)

	if err := denylist.RevokeSession(context.Background(), testSubject, "sess-shared"); err != nil {
		t.Fatalf("revoke session: %v", err)
	}

	stranger := Caller{
		Subject:   "auth0|somebody-else",
		SessionID: "sess-shared",
		IssuedAt:  issued,
	}

	if denylist.Revoked(context.Background(), stranger) {
		t.Fatal("a session id must never revoke the same id under a different subject")
	}
}

func TestSessionIssuedAfterItsRevocationIsAdmitted(t *testing.T) {
	denylist, _ := newDenylist(t)

	if err := denylist.RevokeSession(context.Background(), testSubject, "sess-laptop"); err != nil {
		t.Fatalf("revoke session: %v", err)
	}

	reissued := Caller{
		Subject:   testSubject,
		SessionID: "sess-laptop",
		IssuedAt:  time.Now().Add(time.Minute),
	}

	if denylist.Revoked(context.Background(), reissued) {
		t.Fatal("a token issued after the revocation must be admitted")
	}
}
