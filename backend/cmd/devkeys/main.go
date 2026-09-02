package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
)

func main() {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("SERVICE_TOKEN_SEED=%s\n", base64.StdEncoding.EncodeToString(private.Seed()))
	fmt.Printf("SERVICE_TOKEN_PUBLIC_KEY=%s\n", base64.StdEncoding.EncodeToString(public))
}
