package scan

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	chunkSize   = 32 * 1024
	dialTimeout = 3 * time.Second
	scanTimeout = 30 * time.Second
)

type Clamd struct {
	address string
}

func NewClamd(address string) *Clamd { return &Clamd{address: address} }

func (c *Clamd) Name() string { return ClamAV }

func (c *Clamd) Scan(ctx context.Context, content []byte) (Outcome, error) {
	dialer := net.Dialer{Timeout: dialTimeout}

	connection, err := dialer.DialContext(ctx, "tcp", c.address)
	if err != nil {
		return Outcome{}, fmt.Errorf("reach clamd at %s: %w", c.address, err)
	}
	defer connection.Close()

	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		deadline = time.Now().Add(scanTimeout)
	}
	if err := connection.SetDeadline(deadline); err != nil {
		return Outcome{}, fmt.Errorf("set clamd deadline: %w", err)
	}

	if err := stream(connection, content); err != nil {
		return Outcome{}, err
	}

	return interpret(connection)
}

func stream(connection net.Conn, content []byte) error {
	if _, err := connection.Write([]byte("zINSTREAM\x00")); err != nil {
		return fmt.Errorf("open clamd stream: %w", err)
	}

	header := make([]byte, 4)

	for offset := 0; offset < len(content); offset += chunkSize {
		end := min(offset+chunkSize, len(content))

		binary.BigEndian.PutUint32(header, uint32(end-offset))
		if _, err := connection.Write(header); err != nil {
			return fmt.Errorf("write clamd chunk header: %w", err)
		}
		if _, err := connection.Write(content[offset:end]); err != nil {
			return fmt.Errorf("write clamd chunk: %w", err)
		}
	}

	binary.BigEndian.PutUint32(header, 0)
	if _, err := connection.Write(header); err != nil {
		return fmt.Errorf("close clamd stream: %w", err)
	}

	return nil
}

func interpret(connection net.Conn) (Outcome, error) {
	reply, err := bufio.NewReader(connection).ReadString('\x00')
	if err != nil && reply == "" {
		return Outcome{}, fmt.Errorf("read clamd reply: %w", err)
	}

	reply = strings.TrimSuffix(strings.TrimSpace(reply), "\x00")

	switch {
	case strings.HasSuffix(reply, "OK"):
		return Outcome{Clean: true}, nil
	case strings.HasSuffix(reply, "FOUND"):
		return Outcome{Clean: false, Signature: signatureOf(reply)}, nil
	default:
		return Outcome{}, fmt.Errorf("clamd reported %q", reply)
	}
}

func signatureOf(reply string) string {
	_, remainder, found := strings.Cut(reply, ": ")
	if !found {
		return reply
	}
	return strings.TrimSpace(strings.TrimSuffix(remainder, "FOUND"))
}
