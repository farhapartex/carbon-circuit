package scan

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

type Outcome struct {
	Clean     bool
	Signature string
}

type Scanner interface {
	Name() string
	Scan(ctx context.Context, content []byte) (Outcome, error)
}

const (
	Disabled = "none"
	ClamAV   = "clamav"
)

func Open(kind, address string, logger *slog.Logger) (Scanner, error) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case ClamAV:
		if address == "" {
			return nil, fmt.Errorf("scanner %q requires an address", ClamAV)
		}
		return NewClamd(address), nil
	case Disabled, "":
		logger.Warn("malware scanning is disabled",
			slog.String("consequence",
				"every accepted document records scanned_by=none and was never scanned for malware"))
		return passthrough{}, nil
	default:
		return nil, fmt.Errorf("unknown scanner %q", kind)
	}
}

type passthrough struct{}

func (passthrough) Name() string { return Disabled }

func (passthrough) Scan(context.Context, []byte) (Outcome, error) {
	return Outcome{Clean: true}, nil
}
