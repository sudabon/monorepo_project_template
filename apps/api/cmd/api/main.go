package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/sudabon/monorepo_project_template/apps/api/internal/handler"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/repository"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/usecase"
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
	db, err := database.Open(os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer db.Close()
	address := os.Getenv("HTTP_ADDR")
	if address == "" {
		address = ":8080"
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	slog.Info("API listening", "address", listener.Addr().String())
	return server.Serve(listener, handler.New(usecase.NewItems(repository.NewItems(db)), db.PingContext), wait)
}
func main() {
	slog.SetDefault(logging.New(os.Stdout))
	if err := run(); err != nil {
		slog.Error("API stopped", "error", err)
		os.Exit(1)
	}
}
