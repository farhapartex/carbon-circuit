package reference

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	Length   = 22
	bits     = 128
)

var base = big.NewInt(int64(len(alphabet)))

func New() (string, error) {
	limit := new(big.Int).Lsh(big.NewInt(1), bits)

	value, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return "", fmt.Errorf("draw public reference: %w", err)
	}

	encoded := make([]byte, Length)
	remainder := new(big.Int)

	for index := Length - 1; index >= 0; index-- {
		value.DivMod(value, base, remainder)
		encoded[index] = alphabet[remainder.Int64()]
	}

	return string(encoded), nil
}
