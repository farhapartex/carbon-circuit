package config

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
)

func Ed25519PrivateKey(seed string) (ed25519.PrivateKey, error) {
	decoded, err := base64.StdEncoding.DecodeString(seed)
	if err != nil {
		return nil, fmt.Errorf("decode service token seed: %w", err)
	}

	if len(decoded) != ed25519.SeedSize {
		return nil, fmt.Errorf("service token seed must be %d bytes, got %d", ed25519.SeedSize, len(decoded))
	}

	return ed25519.NewKeyFromSeed(decoded), nil
}

func Ed25519PublicKey(encoded string) (ed25519.PublicKey, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode service token public key: %w", err)
	}

	if len(decoded) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("service token public key must be %d bytes, got %d", ed25519.PublicKeySize, len(decoded))
	}

	return ed25519.PublicKey(decoded), nil
}
