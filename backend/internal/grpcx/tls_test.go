package grpcx_test

import (
	"context"
	"net"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/carboncircuit/backend/internal/grpcx"
)

func issueCertificates(t *testing.T) string {
	t.Helper()

	directory := t.TempDir()
	command := exec.Command("go", "run", "../../cmd/devcerts",
		"-out", directory, "-services", "identity-service,api-gateway")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("generate certificates: %v\n%s", err, output)
	}

	return directory
}

func files(directory, service string) grpcx.TLSFiles {
	return grpcx.TLSFiles{
		CertificateAuthority: filepath.Join(directory, "ca.crt"),
		Certificate:          filepath.Join(directory, service+".crt"),
		PrivateKey:           filepath.Join(directory, service+".key"),
	}
}

func serve(t *testing.T, directory string) string {
	t.Helper()

	credentials, err := grpcx.ServerCredentials(files(directory, "identity-service"))
	if err != nil {
		t.Fatalf("server credentials: %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	server := grpc.NewServer(grpc.Creds(credentials))
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	return listener.Addr().String()
}

func check(t *testing.T, address string, transport grpc.DialOption) error {
	t.Helper()

	connection, err := grpc.NewClient(address, transport)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer connection.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = grpc_health_v1.NewHealthClient(connection).Check(
		ctx, &grpc_health_v1.HealthCheckRequest{},
	)
	return err
}

func TestMutualTLSAdmitsAKnownClient(t *testing.T) {
	directory := issueCertificates(t)
	address := serve(t, directory)

	credentials, err := grpcx.ClientCredentials(files(directory, "api-gateway"), "identity-service")
	if err != nil {
		t.Fatalf("client credentials: %v", err)
	}

	if err := check(t, address, grpc.WithTransportCredentials(credentials)); err != nil {
		t.Fatalf("expected a certificate from the same authority to be admitted: %v", err)
	}
}

func TestMutualTLSRefusesAClientFromAnotherAuthority(t *testing.T) {
	directory := issueCertificates(t)
	address := serve(t, directory)
	foreign := issueCertificates(t)

	credentials, err := grpcx.ClientCredentials(
		grpcx.TLSFiles{
			CertificateAuthority: filepath.Join(directory, "ca.crt"),
			Certificate:          filepath.Join(foreign, "api-gateway.crt"),
			PrivateKey:           filepath.Join(foreign, "api-gateway.key"),
		},
		"identity-service",
	)
	if err != nil {
		t.Fatalf("client credentials: %v", err)
	}

	if err := check(t, address, grpc.WithTransportCredentials(credentials)); err == nil {
		t.Fatal("expected a certificate from a foreign authority to be refused")
	}
}
