package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

type ServerOptions struct {
	Address         string
	Handler         http.Handler
	Logger          *slog.Logger
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

func Serve(ctx context.Context, options ServerOptions) error {
	server := &http.Server{
		Addr:              options.Address,
		Handler:           options.Handler,
		ReadHeaderTimeout: options.ReadTimeout,
		ReadTimeout:       options.ReadTimeout,
		WriteTimeout:      options.WriteTimeout,
	}

	notifyCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	listenFailed := make(chan error, 1)
	go func() {
		options.Logger.Info("listening", slog.String("address", options.Address))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			listenFailed <- err
		}
	}()

	select {
	case err := <-listenFailed:
		return err
	case <-notifyCtx.Done():
		options.Logger.Info("shutdown signal received, draining in-flight requests")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), options.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}

	options.Logger.Info("shutdown complete")
	return nil
}
