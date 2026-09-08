package scan_test

import (
	"bufio"
	"context"
	"encoding/binary"
	"io"
	"net"
	"testing"

	"github.com/carboncircuit/backend/services/evidence-service/internal/scan"
)

func clamdStub(t *testing.T, reply string) (string, func() []byte) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { listener.Close() })

	received := make(chan []byte, 1)

	go func() {
		connection, err := listener.Accept()
		if err != nil {
			return
		}
		defer connection.Close()

		reader := bufio.NewReader(connection)
		command, err := reader.ReadString('\x00')
		if err != nil || command != "zINSTREAM\x00" {
			received <- nil
			return
		}

		var body []byte
		header := make([]byte, 4)

		for {
			if _, err := io.ReadFull(reader, header); err != nil {
				break
			}
			length := binary.BigEndian.Uint32(header)
			if length == 0 {
				break
			}
			chunk := make([]byte, length)
			if _, err := io.ReadFull(reader, chunk); err != nil {
				break
			}
			body = append(body, chunk...)
		}

		received <- body
		connection.Write([]byte(reply))
	}()

	return listener.Addr().String(), func() []byte { return <-received }
}

func TestCleanReplyIsAccepted(t *testing.T) {
	address, streamed := clamdStub(t, "stream: OK\x00")

	payload := []byte("a perfectly ordinary utility statement")

	outcome, err := scan.NewClamd(address).Scan(context.Background(), payload)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if !outcome.Clean {
		t.Fatal("expected a clean outcome")
	}
	if got := string(streamed()); got != string(payload) {
		t.Fatalf("clamd received %q, expected the whole payload", got)
	}
}

func TestSignatureMatchIsRefused(t *testing.T) {
	address, _ := clamdStub(t, "stream: Eicar-Test-Signature FOUND\x00")

	outcome, err := scan.NewClamd(address).Scan(context.Background(), []byte("X5O!P%@AP[4"))
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if outcome.Clean {
		t.Fatal("a signature match must never be reported clean")
	}
	if outcome.Signature != "Eicar-Test-Signature" {
		t.Fatalf("expected the signature name, got %q", outcome.Signature)
	}
}

func TestLargePayloadIsStreamedWhole(t *testing.T) {
	address, streamed := clamdStub(t, "stream: OK\x00")

	payload := make([]byte, 100*1024)
	for index := range payload {
		payload[index] = byte(index % 251)
	}

	if _, err := scan.NewClamd(address).Scan(context.Background(), payload); err != nil {
		t.Fatalf("scan: %v", err)
	}

	received := streamed()
	if len(received) != len(payload) {
		t.Fatalf("clamd received %d bytes of %d", len(received), len(payload))
	}
	for index := range payload {
		if received[index] != payload[index] {
			t.Fatalf("payload corrupted at byte %d", index)
		}
	}
}

func TestUnparseableReplyIsAnError(t *testing.T) {
	address, _ := clamdStub(t, "stream: SIZE LIMIT EXCEEDED ERROR\x00")

	if _, err := scan.NewClamd(address).Scan(context.Background(), []byte("payload")); err == nil {
		t.Fatal("an unrecognised clamd reply must not be treated as clean")
	}
}

func TestUnreachableScannerIsAnError(t *testing.T) {
	if _, err := scan.NewClamd("127.0.0.1:1").Scan(context.Background(), []byte("payload")); err == nil {
		t.Fatal("an unreachable scanner must not be treated as clean")
	}
}
