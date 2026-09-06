package migrations_test

import (
	"context"
	"testing"

	"github.com/sudabon/monorepo_project_template/apps/api/internal/testdb"
	"github.com/sudabon/monorepo_project_template/apps/api/migrations"
)

func TestRollbackAndReapply(t *testing.T) {
	db := testdb.Open(t)
	p, err := migrations.NewProvider(db)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Down(context.Background()); err != nil {
		t.Fatal(err)
	}
	var table *string
	if err := db.QueryRow(`SELECT to_regclass('items')::text`).Scan(&table); err != nil || table != nil {
		t.Fatalf("table after down = %v, %v", table, err)
	}
	if _, err := p.Up(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO items(name) VALUES ('reapplied')`); err != nil {
		t.Fatal(err)
	}
}
