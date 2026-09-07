package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sudabon/monorepo_project_template/apps/api/migrations"
	"github.com/sudabon/monorepo_project_template/packages/go-platform/config"
	"github.com/sudabon/monorepo_project_template/packages/go-platform/database"
	"github.com/sudabon/monorepo_project_template/packages/go-platform/migrate"
)

func run() error {
	arg := ""
	if len(os.Args) == 2 {
		arg = os.Args[1]
	}
	databaseURL, err := config.Require("DATABASE_URL")
	if err != nil {
		return err
	}
	db, err := database.Open(databaseURL)
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
	return migrate.Run(ctx, p, arg)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
