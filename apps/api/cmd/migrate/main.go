package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sudabon/monorepo_project_template/apps/api/migrations"
	"github.com/sudabon/monorepo_project_template/packages/go-platform/database"
)

func run() error {
	if len(os.Args) != 2 || (os.Args[1] != "up" && os.Args[1] != "down") {
		return fmt.Errorf("usage: migrate <up|down>")
	}
	if os.Getenv("DATABASE_URL") == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	db, err := database.Open(os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer db.Close()
	p, err := migrations.NewProvider(db)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if os.Args[1] == "up" {
		_, err = p.Up(ctx)
	} else {
		_, err = p.Down(ctx)
	}
	return err
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
