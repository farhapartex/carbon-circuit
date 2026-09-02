package grpcx

import (
	"context"
	"log/slog"
	"net"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type ServerOptions struct {
	Address         string
	Logger          *slog.Logger
	ShutdownTimeout time.Duration
	Register        func(*grpc.Server)
	ReportHealth    func(context.Context) bool
	HealthInterval  time.Duration
	ServiceName     string
	Interceptors    []grpc.UnaryServerInterceptor
	TransportCreds  credentials.TransportCredentials
}

func Serve(ctx context.Context, options ServerOptions) error {
	listener, err := net.Listen("tcp", options.Address)
	if err != nil {
		return err
	}

	interceptors := append([]grpc.UnaryServerInterceptor{
		CorrelateUnary(),
		RecoverUnary(options.Logger),
		LogUnary(options.Logger),
	}, options.Interceptors...)

	settings := []grpc.ServerOption{grpc.ChainUnaryInterceptor(interceptors...)}
	if options.TransportCreds != nil {
		settings = append(settings, grpc.Creds(options.TransportCreds))
	}

	server := grpc.NewServer(settings...)

	options.Register(server)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	reflection.Register(server)

	notifyCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go watchHealth(notifyCtx, options, healthServer)

	serveFailed := make(chan error, 1)
	go func() {
		options.Logger.Info("listening", slog.String("address", options.Address))
		if serveErr := server.Serve(listener); serveErr != nil {
			serveFailed <- serveErr
		}
	}()

	select {
	case err := <-serveFailed:
		return err
	case <-notifyCtx.Done():
		options.Logger.Info("shutdown signal received, draining in-flight calls")
	}

	healthServer.SetServingStatus(options.ServiceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	stopped := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		options.Logger.Info("shutdown complete")
	case <-time.After(options.ShutdownTimeout):
		server.Stop()
		options.Logger.Warn("shutdown deadline exceeded, forced stop")
	}

	return nil
}

func watchHealth(ctx context.Context, options ServerOptions, healthServer *health.Server) {
	apply := func() {
		status := grpc_health_v1.HealthCheckResponse_NOT_SERVING
		if options.ReportHealth(ctx) {
			status = grpc_health_v1.HealthCheckResponse_SERVING
		}
		healthServer.SetServingStatus(options.ServiceName, status)
		healthServer.SetServingStatus("", status)
	}

	apply()

	ticker := time.NewTicker(options.HealthInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			apply()
		}
	}
}
