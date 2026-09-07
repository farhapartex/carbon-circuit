package httpx

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"

	"github.com/carboncircuit/backend/internal/cache"
	"github.com/carboncircuit/backend/internal/ratelimit"
)

func limitedRouter(t *testing.T, trustProxies []string) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	server := miniredis.RunT(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := cache.New(server.Addr(), "", 0, logger)

	limiter, err := ratelimit.New(client.Redis(), "ratelimit", []ratelimit.Rule{
		{
			Name:      "public_ip",
			PerMinute: 60,
			Burst:     2,
			KeyFunc:   func(r ratelimit.Request) string { return "ip:" + r.ClientIP },
			AppliesTo: func(r ratelimit.Request) bool { return true },
		},
	})
	if err != nil {
		t.Fatalf("build limiter: %v", err)
	}

	router := gin.New()
	if err := router.SetTrustedProxies(trustProxies); err != nil {
		t.Fatalf("set trusted proxies: %v", err)
	}

	router.Use(Correlate(), EndpointClass("public_read"), RateLimit(limiter, logger))
	router.GET("/thing", func(c *gin.Context) {
		Data(c, http.StatusOK, gin.H{"ok": true})
	})

	return router
}

func sendForwarded(router *gin.Engine, forwardedFor string) int {
	request := httptest.NewRequest(http.MethodGet, "/thing", nil)
	request.RemoteAddr = "192.0.2.10:5555"
	if forwardedFor != "" {
		request.Header.Set("X-Forwarded-For", forwardedFor)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder.Code
}

func TestRotatingForwardedHeadersCannotEvadeTheLimit(t *testing.T) {
	router := limitedRouter(t, []string{"0.0.0.0/0"})

	throttled := false
	for attempt := 0; attempt < 40; attempt++ {
		if sendForwarded(router, forwardedAddress(attempt)) == http.StatusTooManyRequests {
			throttled = true
			break
		}
	}

	if !throttled {
		t.Fatal("a caller rotating X-Forwarded-For evaded the per-address limit")
	}
}

func TestLimitStillAppliesWithNoForwardedHeader(t *testing.T) {
	router := limitedRouter(t, nil)

	throttled := false
	for attempt := 0; attempt < 40; attempt++ {
		if sendForwarded(router, "") == http.StatusTooManyRequests {
			throttled = true
			break
		}
	}

	if !throttled {
		t.Fatal("expected the peer address to be limited")
	}
}

func forwardedAddress(attempt int) string {
	return fmt.Sprintf("203.0.%d.%d", attempt/250, attempt%250)
}
