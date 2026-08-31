package httpx

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/carboncircuit/backend/internal/auth"
)

type stubVerifier struct {
	caller auth.Caller
	err    error
}

func (s stubVerifier) Verify(context.Context, string) (auth.Caller, error) {
	return s.caller, s.err
}

type stubDenylist struct {
	revoked bool
}

func (s stubDenylist) Revoked(context.Context, auth.Caller) bool { return s.revoked }

func protectedRouter(verifier TokenVerifier, denylist RevocationChecker) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Correlate())
	router.GET("/protected",
		Authenticate(verifier, denylist, slog.New(slog.NewTextHandler(io.Discard, nil))),
		func(c *gin.Context) {
			caller, verified := auth.CallerFrom(c.Request.Context())
			if !verified {
				Fail(c, CodeInternal)
				return
			}
			Data(c, http.StatusOK, gin.H{"subject": caller.Subject})
		},
	)
	return router
}

func call(router *gin.Engine, authorization string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func errorCode(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	return body.Error.Code
}

func TestAuthenticateRejectsUnusableAuthorizationHeaders(t *testing.T) {
	router := protectedRouter(stubVerifier{caller: auth.Caller{Subject: "auth0|ok"}}, stubDenylist{})

	headers := map[string]string{
		"absent":          "",
		"empty bearer":    "Bearer ",
		"basic scheme":    "Basic dXNlcjpwYXNz",
		"scheme only":     "Bearer",
		"no scheme":       "eyJhbGciOiJSUzI1NiJ9.e30.sig",
		"bearer no gap":   "Bearereyj",
		"whitespace only": "Bearer    ",
	}

	for name, header := range headers {
		t.Run(name, func(t *testing.T) {
			recorder := call(router, header)

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", recorder.Code)
			}
			if code := errorCode(t, recorder); code != string(CodeUnauthenticated) {
				t.Fatalf("expected UNAUTHENTICATED, got %s", code)
			}
		})
	}
}

func TestAuthenticateAdmitsVerifiedCaller(t *testing.T) {
	verifier := stubVerifier{caller: auth.Caller{Subject: "auth0|abc", IssuedAt: time.Now()}}
	recorder := call(protectedRouter(verifier, stubDenylist{}), "Bearer any-token")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var body struct {
		Data struct {
			Subject string `json:"subject"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	if body.Data.Subject != "auth0|abc" {
		t.Fatalf("expected verified subject to reach the handler, got %q", body.Data.Subject)
	}
}

func TestAuthenticateRejectsInvalidToken(t *testing.T) {
	verifier := stubVerifier{err: auth.ErrInvalidToken}
	recorder := call(protectedRouter(verifier, stubDenylist{}), "Bearer forged")

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
	if code := errorCode(t, recorder); code != string(CodeUnauthenticated) {
		t.Fatalf("expected UNAUTHENTICATED, got %s", code)
	}
}

func TestAuthenticateReportsUnreachableKeysAsDependencyFailure(t *testing.T) {
	verifier := stubVerifier{err: auth.ErrKeysUnavailable}
	recorder := call(protectedRouter(verifier, stubDenylist{}), "Bearer genuine")

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}
	if code := errorCode(t, recorder); code != string(CodeDependencyUnavailable) {
		t.Fatalf("expected DEPENDENCY_UNAVAILABLE, got %s", code)
	}
}

func TestAuthenticateRejectsRevokedCaller(t *testing.T) {
	verifier := stubVerifier{caller: auth.Caller{Subject: "auth0|revoked", IssuedAt: time.Now()}}
	recorder := call(protectedRouter(verifier, stubDenylist{revoked: true}), "Bearer genuine")

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
	if code := errorCode(t, recorder); code != string(CodeTokenRevoked) {
		t.Fatalf("expected TOKEN_REVOKED, got %s", code)
	}
}
