package session

import (
	"errors"
	"testing"
	"time"
)

func TestMemoryCreateGetDelete(t *testing.T) {
	store := NewMemory()
	exerciseStore(t, store)
}

func TestMemoryRejectsExpiredAndUnknown(t *testing.T) {
	now := time.Now()
	store := NewMemory(WithIdle(50*time.Millisecond), WithAbsolute(time.Second), WithNow(func() time.Time { return now }))
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

func TestMemorySlidingRefreshStaysWithinAbsolute(t *testing.T) {
	now := time.Now()
	store := NewMemory(WithIdle(100*time.Millisecond), WithAbsolute(250*time.Millisecond), WithNow(func() time.Time { return now }))
	sess, err := store.Create(t.Context(), "u1", "Demo")
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(80 * time.Millisecond)
	got, err := store.Get(t.Context(), sess.ID)
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(80 * time.Millisecond)
	if _, err := store.Get(t.Context(), sess.ID); err != nil {
		t.Fatalf("sliding should still be valid: %v", err)
	}
	now = got.CreatedAt.Add(260 * time.Millisecond)
	if _, err := store.Get(t.Context(), sess.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("absolute expiry err = %v", err)
	}
}

func exerciseStore(t *testing.T, store Store) {
	t.Helper()
	sess, err := store.Create(t.Context(), "user-1", "Demo")
	if err != nil {
		t.Fatal(err)
	}
	if sess.ID == "" || sess.CSRFToken == "" || sess.ID == sess.CSRFToken || sess.UserID != "user-1" || sess.Name != "Demo" {
		t.Fatalf("session = %+v", sess)
	}
	got, err := store.Get(t.Context(), sess.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != sess.ID || got.CSRFToken != sess.CSRFToken || got.UserID != sess.UserID {
		t.Fatalf("got = %+v", got)
	}
	if err := store.Delete(t.Context(), sess.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(t.Context(), sess.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("after delete err = %v", err)
	}
}
