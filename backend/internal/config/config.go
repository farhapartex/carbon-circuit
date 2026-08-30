package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Loader struct {
	prefixHint string
	missing    []string
	invalid    []string
}

func NewLoader(prefixHint string) *Loader {
	return &Loader{prefixHint: prefixHint}
}

func (l *Loader) String(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		l.missing = append(l.missing, key)
	}
	return value
}

func (l *Loader) StringDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func (l *Loader) Int(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		l.invalid = append(l.invalid, fmt.Sprintf("%s must be an integer, got %q", key, raw))
		return fallback
	}
	return parsed
}

func (l *Loader) Duration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		l.invalid = append(l.invalid, fmt.Sprintf("%s must be a duration such as 30s, got %q", key, raw))
		return fallback
	}
	return parsed
}

func (l *Loader) Err() error {
	if len(l.missing) == 0 && len(l.invalid) == 0 {
		return nil
	}

	var report strings.Builder
	report.WriteString(l.prefixHint)
	report.WriteString(" configuration is invalid.")
	for _, key := range l.missing {
		report.WriteString("\n  missing: ")
		report.WriteString(key)
	}
	for _, problem := range l.invalid {
		report.WriteString("\n  invalid: ")
		report.WriteString(problem)
	}
	return fmt.Errorf("%s", report.String())
}
