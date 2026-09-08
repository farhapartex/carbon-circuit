package handler

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	serverReadTimeout = 300 * time.Millisecond
	trickleDuration   = 1200 * time.Millisecond
)

func trickleServer(t *testing.T, handlers *Handlers, widen bool) string {
	t.Helper()

	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/upload", func(c *gin.Context) {
		if widen {
			if err := handlers.widenUploadWindow(c); err != nil {
				c.String(http.StatusInternalServerError, "widen: %v", err)
				return
			}
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.String(http.StatusRequestTimeout, "read: %v", err)
			return
		}

		c.String(http.StatusOK, "%d", len(body))
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	server := &http.Server{Handler: router, ReadTimeout: serverReadTimeout}
	go server.Serve(listener)
	t.Cleanup(func() { server.Close() })

	return listener.Addr().String()
}

func trickleUpload(t *testing.T, address string, chunks int) (string, error) {
	t.Helper()

	connection, err := net.Dial("tcp", address)
	if err != nil {
		return "", err
	}
	defer connection.Close()

	if err := connection.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return "", err
	}

	payload := make([]byte, 64)
	for index := range payload {
		payload[index] = 'a'
	}

	request := fmt.Sprintf(
		"POST /upload HTTP/1.1\r\nHost: probe\r\nContent-Length: %d\r\n\r\n",
		len(payload)*chunks,
	)
	if _, err := connection.Write([]byte(request)); err != nil {
		return "", err
	}

	pause := trickleDuration / time.Duration(chunks)

	for sent := 0; sent < chunks; sent++ {
		if _, err := connection.Write(payload); err != nil {
			return "", err
		}
		time.Sleep(pause)
	}

	reply, err := io.ReadAll(connection)
	return string(reply), err
}

func probeHandlers(window time.Duration) *Handlers {
	return &Handlers{
		Logger:                slog.New(slog.NewTextHandler(io.Discard, nil)),
		EvidenceUploadWindow:  window,
		EvidenceUploadTimeout: window,
	}
}

func TestASlowBodyIsCutOffWithoutAWiderWindow(t *testing.T) {
	address := trickleServer(t, probeHandlers(5*time.Second), false)

	reply, err := trickleUpload(t, address, 8)
	if err != nil {
		return
	}

	if isSuccessful(reply) {
		t.Fatalf("a body slower than the %s read timeout should not have completed, got %q",
			serverReadTimeout, firstLine(reply))
	}
}

func TestAWiderWindowLetsASlowBodyFinish(t *testing.T) {
	address := trickleServer(t, probeHandlers(5*time.Second), true)

	reply, err := trickleUpload(t, address, 8)
	if err != nil {
		t.Fatalf("trickled upload: %v", err)
	}

	if !isSuccessful(reply) {
		t.Fatalf("the widened window should have carried the slow body through, got %q",
			firstLine(reply))
	}
}

func TestAZeroWindowLeavesTheServerTimeoutInPlace(t *testing.T) {
	address := trickleServer(t, probeHandlers(0), true)

	reply, err := trickleUpload(t, address, 8)
	if err != nil {
		return
	}

	if isSuccessful(reply) {
		t.Fatal("a zero window must not widen anything, so the server timeout should still apply")
	}
}

func isSuccessful(reply string) bool {
	return len(reply) >= 12 && reply[9:12] == "200"
}

func firstLine(reply string) string {
	for index, character := range reply {
		if character == '\r' || character == '\n' {
			return reply[:index]
		}
	}
	return reply
}
