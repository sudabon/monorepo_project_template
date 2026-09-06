package main

import (
	"cmp"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"

	"github.com/sudabon/monorepo_project_template/apps/bff/internal/handler"
	"github.com/sudabon/monorepo_project_template/apps/bff/internal/identity"
	"github.com/sudabon/monorepo_project_template/apps/bff/internal/session"
	"github.com/sudabon/monorepo_project_template/packages/go-platform/database"
	"github.com/sudabon/monorepo_project_template/packages/go-platform/logging"
	"github.com/sudabon/monorepo_project_template/packages/go-platform/server"
)

func run() error {
	wait, err := server.ShutdownTimeout()
	if err != nil {
		return err
	}
	if os.Getenv("DATABASE_URL") == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		return fmt.Errorf("BACKEND_URL is required")
	}
	backend, err := url.Parse(backendURL)
	if err != nil || backend.Scheme == "" || backend.Host == "" {
		return fmt.Errorf("BACKEND_URL must be an absolute URL")
	}
	username := os.Getenv("BFF_DEMO_USERNAME")
	password := os.Getenv("BFF_DEMO_PASSWORD")
	if username == "" || password == "" {
		return fmt.Errorf("BFF_DEMO_USERNAME and BFF_DEMO_PASSWORD are required")
	}
	db, err := database.Open(os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer db.Close()
	address := cmp.Or(os.Getenv("HTTP_ADDR"), ":8081")
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	secure := os.Getenv("BFF_COOKIE_SECURE") != "false" && os.Getenv("BFF_COOKIE_SECURE") != "0"
	slog.Info("BFF listening", "address", listener.Addr().String())
	return server.Serve(listener, handler.New(handler.Deps{
		Store:        session.NewPostgres(db),
		Users:        identity.Static{Username: username, Password: password, User: identity.User{ID: username, Name: username}},
		Backend:      backend,
		CookieSecure: secure,
		Ping:         db.PingContext,
	}), wait)
}

func main() {
	slog.SetDefault(logging.New(os.Stdout))
	if err := run(); err != nil {
		slog.Error("BFF stopped", "error", err)
		os.Exit(1)
	}
}
