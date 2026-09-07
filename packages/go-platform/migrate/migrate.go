// Package migrate runs goose up/down from a shared CLI argument parser.
package migrate

import (
	"context"
	"fmt"

	"github.com/pressly/goose/v3"
)

type Provider interface {
	Up(ctx context.Context) ([]*goose.MigrationResult, error)
	Down(ctx context.Context) (*goose.MigrationResult, error)
}

func Run(ctx context.Context, provider Provider, arg string) error {
	switch arg {
	case "up":
		_, err := provider.Up(ctx)
		return err
	case "down":
		_, err := provider.Down(ctx)
		return err
	default:
		return fmt.Errorf("usage: migrate <up|down>")
	}
}
