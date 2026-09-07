package apikey

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

const (
	Scheme       = "cc_live"
	PrefixLength = 8
	secretBytes  = 32

	prefixAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
)

var (
	ErrMalformed   = errors.New("api key is not in the expected form")
	ErrPepperUnset = errors.New("api key pepper is not configured")
)

type Issued struct {
	Prefix string
	Secret string
	HMAC   []byte
}

func (i Issued) Presented() string {
	return Scheme + "_" + i.Prefix + "_" + i.Secret
}

type Hasher struct {
	pepper []byte
}

func NewHasher(pepper string) (*Hasher, error) {
	if strings.TrimSpace(pepper) == "" {
		return nil, ErrPepperUnset
	}
	return &Hasher{pepper: []byte(pepper)}, nil
}

func (h *Hasher) Issue() (Issued, error) {
	prefix, err := randomPrefix()
	if err != nil {
		return Issued{}, err
	}

	raw := make([]byte, secretBytes)
	if _, err := rand.Read(raw); err != nil {
		return Issued{}, fmt.Errorf("draw api key secret: %w", err)
	}

	secret := base64.RawURLEncoding.EncodeToString(raw)

	return Issued{
		Prefix: prefix,
		Secret: secret,
		HMAC:   h.sign(secret),
	}, nil
}

func (h *Hasher) Matches(secret string, expected []byte) bool {
	return subtle.ConstantTimeCompare(h.sign(secret), expected) == 1
}

func (h *Hasher) sign(secret string) []byte {
	mac := hmac.New(sha256.New, h.pepper)
	mac.Write([]byte(secret))
	return mac.Sum(nil)
}

type Presented struct {
	Prefix string
	Secret string
}

func Parse(presented string) (Presented, error) {
	parts := strings.SplitN(presented, "_", 4)
	if len(parts) != 4 {
		return Presented{}, ErrMalformed
	}

	if parts[0]+"_"+parts[1] != Scheme {
		return Presented{}, ErrMalformed
	}

	if len(parts[2]) != PrefixLength || parts[3] == "" {
		return Presented{}, ErrMalformed
	}

	return Presented{Prefix: parts[2], Secret: parts[3]}, nil
}

func randomPrefix() (string, error) {
	drawn := make([]byte, PrefixLength)
	if _, err := rand.Read(drawn); err != nil {
		return "", fmt.Errorf("draw api key prefix: %w", err)
	}

	prefix := make([]byte, PrefixLength)
	for index, value := range drawn {
		prefix[index] = prefixAlphabet[int(value)%len(prefixAlphabet)]
	}

	return string(prefix), nil
}
