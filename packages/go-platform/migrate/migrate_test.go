package migrate

import (
	"context"
	"errors"
	"testing"

	"github.com/pressly/goose/v3"
)

type stub struct {
	ups, downs int
	err        error
}

func (s *stub) Up(context.Context) ([]*goose.MigrationResult, error) {
	s.ups++
	return nil, s.err
}

func (s *stub) Down(context.Context) (*goose.MigrationResult, error) {
	s.downs++
	return nil, s.err
}

func TestRunUpDownAndRejectsUnknownArg(t *testing.T) {
	ctx := t.Context()
	p := &stub{}
	if err := Run(ctx, p, "up"); err != nil || p.ups != 1 || p.downs != 0 {
		t.Fatalf("up = %d/%d, %v", p.ups, p.downs, err)
	}
	if err := Run(ctx, p, "down"); err != nil || p.ups != 1 || p.downs != 1 {
		t.Fatalf("down = %d/%d, %v", p.ups, p.downs, err)
	}
	if err := Run(ctx, p, "sideways"); err == nil || p.ups != 1 || p.downs != 1 {
		t.Fatalf("invalid = %d/%d, %v", p.ups, p.downs, err)
	}
	p.err = errors.New("failed")
	if err := Run(ctx, p, "up"); !errors.Is(err, p.err) {
		t.Fatalf("up error = %v", err)
	}
}
