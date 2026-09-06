package session

import (
	"errors"
	"testing"
	"time"

	"github.com/sudabon/monorepo_project_template/apps/bff/internal/testdb"
)

func TestPostgresCreateGetDelete(t *testing.T) {
	exerciseStore(t, NewPostgres(testdb.Open(t)))
}

func TestPostgresRejectsExpiredAndUnknown(t *testing.T) {
	now := time.Now()
	store := NewPostgres(testdb.Open(t), WithIdle(50*time.Millisecond), WithAbsolute(time.Second), WithNow(func() time.Time { return now }))
	sess, err := store.Create(t.Context(), "u1", "Demo")
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(60 * time.Millisecond)
	if _, err := store.Get(t.Context(), sess.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expired err = %v", err)
	}
	if _, err := store.Get(t.Context(), "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown err = %v", err)
	}
}
