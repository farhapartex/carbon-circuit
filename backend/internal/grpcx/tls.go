package grpcx

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

type TLSFiles struct {
	CertificateAuthority string
	Certificate          string
	PrivateKey           string
}

func (f TLSFiles) Configured() bool {
	return f.CertificateAuthority != "" && f.Certificate != "" && f.PrivateKey != ""
}

func (f TLSFiles) load() (*x509.CertPool, tls.Certificate, error) {
	authority, err := os.ReadFile(f.CertificateAuthority)
	if err != nil {
		return nil, tls.Certificate{}, fmt.Errorf("read certificate authority: %w", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(authority) {
		return nil, tls.Certificate{}, fmt.Errorf("certificate authority %s is not valid pem", f.CertificateAuthority)
	}

	pair, err := tls.LoadX509KeyPair(f.Certificate, f.PrivateKey)
	if err != nil {
		return nil, tls.Certificate{}, fmt.Errorf("load key pair: %w", err)
	}

	return pool, pair, nil
}

func ServerCredentials(files TLSFiles) (credentials.TransportCredentials, error) {
	pool, pair, err := files.load()
	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{pair},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
		MinVersion:   tls.VersionTLS13,
	}), nil
}

func ClientCredentials(files TLSFiles, serverName string) (credentials.TransportCredentials, error) {
	pool, pair, err := files.load()
	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{pair},
		RootCAs:      pool,
		ServerName:   serverName,
		MinVersion:   tls.VersionTLS13,
	}), nil
}
