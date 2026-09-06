package session

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Postgres struct {
	db *sql.DB
	clock
}

func NewPostgres(db *sql.DB, opts ...Option) *Postgres {
	return &Postgres{db: db, clock: newClock(opts)}
}

func (p *Postgres) Create(ctx context.Context, userID, name string) (Session, error) {
	id, err := newID()
	if err != nil {
		return Session{}, err
	}
	csrf, err := newID()
	if err != nil {
		return Session{}, err
	}
	now := p.now()
	sess := Session{ID: id, UserID: userID, Name: name, CSRFToken: csrf, CreatedAt: now, ExpiresAt: p.expireAt(now)}
	_, err = p.db.ExecContext(ctx, `INSERT INTO sessions (id, user_id, display_name, csrf_token, created_at, expires_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		sess.ID, sess.UserID, sess.Name, sess.CSRFToken, sess.CreatedAt, sess.ExpiresAt)
	if err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}
	return sess, nil
}

// TODO(template): periodically delete expired sessions; Get already treats them as missing.

func (p *Postgres) Get(ctx context.Context, id string) (Session, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return Session{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var sess Session
	err = tx.QueryRowContext(ctx, `SELECT id, user_id, display_name, csrf_token, created_at, expires_at FROM sessions WHERE id = $1`, id).
		Scan(&sess.ID, &sess.UserID, &sess.Name, &sess.CSRFToken, &sess.CreatedAt, &sess.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("get session: %w", err)
	}
	if !p.now().Before(sess.ExpiresAt) {
		if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE id = $1`, id); err != nil {
			return Session{}, err
		}
		if err := tx.Commit(); err != nil {
			return Session{}, err
		}
		return Session{}, ErrNotFound
	}
	sess.ExpiresAt = p.expireAt(sess.CreatedAt)
	if !p.now().Before(sess.ExpiresAt) {
		if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE id = $1`, id); err != nil {
			return Session{}, err
		}
		if err := tx.Commit(); err != nil {
			return Session{}, err
		}
		return Session{}, ErrNotFound
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sessions SET expires_at = $2 WHERE id = $1`, id, sess.ExpiresAt); err != nil {
		return Session{}, err
	}
	if err := tx.Commit(); err != nil {
		return Session{}, err
	}
	return sess, nil
}

func (p *Postgres) Delete(ctx context.Context, id string) error {
	_, err := p.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

var _ Store = (*Postgres)(nil)
