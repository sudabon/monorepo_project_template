package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func ShutdownTimeout() (time.Duration, error) {
	wait := 20 * time.Second
	if value := os.Getenv("SHUTDOWN_TIMEOUT"); value != "" {
		var err error
		wait, err = time.ParseDuration(value)
		if err != nil || wait <= 0 {
			return 0, fmt.Errorf("SHUTDOWN_TIMEOUT must be a positive duration")
		}
	}
	return wait, nil
}

func Serve(listener net.Listener, handler http.Handler, wait time.Duration) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 35 * time.Second, IdleTimeout: 60 * time.Second}
	result := make(chan error, 1)
	go func() { result <- srv.Serve(listener) }()
	select {
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		slog.Info("shutdown started", "timeout", wait.String())
	}
	shutdown, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	if err := srv.Shutdown(shutdown); err != nil {
		slog.Warn("shutdown deadline reached", "error", err)
		if closeErr := srv.Close(); closeErr != nil {
			return fmt.Errorf("force close server: %w", closeErr)
		}
		return nil
	}
	slog.Info("shutdown completed")
	return nil
}
