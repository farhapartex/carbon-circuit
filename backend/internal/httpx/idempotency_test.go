package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func idempotentRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Correlate(), RequireIdempotencyKey())

	handler := func(c *gin.Context) {
		key, present := IdempotencyKeyFrom(c.Request.Context())
		Data(c, http.StatusOK, gin.H{"key": key, "present": present})
	}

	router.GET("/thing", handler)
	router.POST("/thing", handler)
	router.PATCH("/thing", handler)
	router.DELETE("/thing", handler)

	return router
}

func send(router *gin.Engine, method, key string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, "/thing", nil)
	if key != "" {
		request.Header.Set("Idempotency-Key", key)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestReadsDoNotRequireAnIdempotencyKey(t *testing.T) {
	recorder := send(idempotentRouter(), http.MethodGet, "")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected reads to pass without a key, got %d", recorder.Code)
	}
}

func TestMutationsWithoutAKeyAreRefused(t *testing.T) {
	router := idempotentRouter()

	for _, method := range []string{http.MethodPost, http.MethodPatch, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			recorder := send(router, method, "")

			if recorder.Code != http.StatusUnprocessableEntity {
				t.Fatalf("expected 422, got %d", recorder.Code)
			}
			if code := errorCode(t, recorder); code != string(CodeIdempotencyKeyRequired) {
				t.Fatalf("expected IDEMPOTENCY_KEY_REQUIRED, got %s", code)
			}
		})
	}
}

func TestUnusableKeysAreRejected(t *testing.T) {
	router := idempotentRouter()

	keys := map[string]string{
		"too short":      "abc",
		"too long":       strings.Repeat("k", 256),
		"embedded tab":   "abc\tdefgh",
		"control char":   "abcdefg\x00",
		"non ascii":      "abcdefgé",
		"internal space": "abcd efgh",
	}

	for name, key := range keys {
		t.Run(name, func(t *testing.T) {
			recorder := send(router, http.MethodPost, key)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", recorder.Code)
			}
			if code := errorCode(t, recorder); code != string(CodeValidation) {
				t.Fatalf("expected VALIDATION_ERROR, got %s", code)
			}
		})
	}
}

func TestUsableKeyReachesTheHandler(t *testing.T) {
	recorder := send(idempotentRouter(), http.MethodPost, "01JD8Z9K2QW4RTY6")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "01JD8Z9K2QW4RTY6") {
		t.Fatalf("expected the key on the request context, got %s", recorder.Body.String())
	}
}
