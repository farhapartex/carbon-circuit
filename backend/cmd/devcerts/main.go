package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const validity = 365 * 24 * time.Hour

func main() {
	directory := flag.String("out", "certs", "directory to write certificates into")
	names := flag.String("services", "api-gateway,identity-service,billing-service", "comma separated service names")
	flag.Parse()

	if err := generate(*directory, strings.Split(*names, ",")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func generate(directory string, services []string) error {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", directory, err)
	}

	authority, authorityKey, err := createAuthority()
	if err != nil {
		return err
	}

	if err := writePair(directory, "ca", authority, authorityKey, nil, nil); err != nil {
		return err
	}

	parsed, err := x509.ParseCertificate(authority)
	if err != nil {
		return fmt.Errorf("parse authority: %w", err)
	}

	for _, service := range services {
		service = strings.TrimSpace(service)
		if service == "" {
			continue
		}

		certificate, key, leafErr := createLeaf(service, parsed, authorityKey)
		if leafErr != nil {
			return leafErr
		}
		if err := writePair(directory, service, certificate, key, nil, nil); err != nil {
			return err
		}
		fmt.Printf("issued %s\n", service)
	}

	return nil
}

func createAuthority() ([]byte, ed25519.PrivateKey, error) {
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate authority key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "CarbonCircuit Development CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(validity),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certificate, err := x509.CreateCertificate(rand.Reader, template, template, key.Public(), key)
	if err != nil {
		return nil, nil, fmt.Errorf("create authority certificate: %w", err)
	}

	return certificate, key, nil
}

func createLeaf(
	service string,
	authority *x509.Certificate,
	authorityKey ed25519.PrivateKey,
) ([]byte, ed25519.PrivateKey, error) {
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate %s key: %w", service, err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 96))
	if err != nil {
		return nil, nil, fmt.Errorf("generate serial: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: service},
		DNSNames:     []string{service, "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(validity),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	certificate, err := x509.CreateCertificate(rand.Reader, template, authority, key.Public(), authorityKey)
	if err != nil {
		return nil, nil, fmt.Errorf("create %s certificate: %w", service, err)
	}

	return certificate, key, nil
}

func writePair(
	directory, name string,
	certificate []byte,
	key ed25519.PrivateKey,
	_ any, _ any,
) error {
	certPath := filepath.Join(directory, name+".crt")
	keyPath := filepath.Join(directory, name+".key")

	if err := writePem(certPath, "CERTIFICATE", certificate, 0o644); err != nil {
		return err
	}

	encoded, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("marshal %s key: %w", name, err)
	}

	return writePem(keyPath, "PRIVATE KEY", encoded, 0o600)
}

func writePem(path, blockType string, contents []byte, mode os.FileMode) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	if err := pem.Encode(file, &pem.Block{Type: blockType, Bytes: contents}); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}
