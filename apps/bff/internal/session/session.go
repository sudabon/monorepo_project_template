package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

const CookieName = "session_id"

const (
	IdleTTL     = 12 * time.Hour
	AbsoluteTTL = 24 * time.Hour
)

var ErrNotFound = errors.New("session not found")

type Session struct {
	ID        string
	UserID    string
	Name      string
	CSRFToken string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type Store interface {
	Create(ctx context.Context, userID, name string) (Session, error)
	Get(ctx context.Context, id string) (Session, error)
	Delete(ctx context.Context, id string) error
}

type clock struct {
	idle, absolute time.Duration
	now            func() time.Time
}

func newClock(opts []Option) clock {
	c := clock{idle: IdleTTL, absolute: AbsoluteTTL, now: time.Now}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

type Option func(*clock)

func WithIdle(d time.Duration) Option     { return func(c *clock) { c.idle = d } }
func WithAbsolute(d time.Duration) Option { return func(c *clock) { c.absolute = d } }
func WithNow(now func() time.Time) Option { return func(c *clock) { c.now = now } }

func (c clock) expireAt(created time.Time) time.Time {
	idle := c.now().Add(c.idle)
	abs := created.Add(c.absolute)
	if idle.After(abs) {
		return abs
	}
	return idle
}

func newID() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

type ctxKey struct{}

func WithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, ctxKey{}, s)
}

func FromContext(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(ctxKey{}).(Session)
	return s, ok
}
