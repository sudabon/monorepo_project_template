package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"

	"github.com/sudabon/monorepo_project_template/apps/bff/internal/handler"
	"github.com/sudabon/monorepo_project_template/apps/bff/internal/identity"
	"github.com/sudabon/monorepo_project_template/apps/bff/internal/session"
	"github.com/sudabon/monorepo_project_template/packages/go-platform/config"
	"github.com/sudabon/monorepo_project_template/packages/go-platform/database"
	"github.com/sudabon/monorepo_project_template/packages/go-platform/logging"
	"github.com/sudabon/monorepo_project_template/packages/go-platform/server"
)

func run() error {
	wait, err := server.ShutdownTimeout()
	if err != nil {
		return err
	}
	databaseURL, err := config.Require("DATABASE_URL")
	if err != nil {
		return err
	}
	backendURL, err := config.Require("BACKEND_URL")
	if err != nil {
		return err
	}
	backend, err := url.Parse(backendURL)
	if err != nil || backend.Scheme == "" || backend.Host == "" {
		return fmt.Errorf("BACKEND_URL must be an absolute URL")
	}
	username, err := config.Require("BFF_DEMO_USERNAME")
	if err != nil {
		return err
	}
	password, err := config.Require("BFF_DEMO_PASSWORD")
	if err != nil {
		return err
	}
	secure, err := config.Bool("BFF_COOKIE_SECURE", true)
	if err != nil {
		return err
	}
	db, err := database.Open(databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	address := config.Or("HTTP_ADDR", ":8081")
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
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
