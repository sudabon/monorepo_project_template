package migrations_test

import (
	"testing"

	"github.com/sudabon/monorepo_project_template/apps/bff/internal/testdb"
)

func TestSessionMigrationApplies(t *testing.T) {
	db := testdb.Open(t)
	var n int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM sessions`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("sessions = %d", n)
	}
}
