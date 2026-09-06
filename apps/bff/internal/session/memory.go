package session

import (
	"context"
	"sync"
)

type Memory struct {
	mu       sync.Mutex
	sessions map[string]Session
	clock
}

func NewMemory(opts ...Option) *Memory {
	return &Memory{sessions: map[string]Session{}, clock: newClock(opts)}
}

func (m *Memory) Create(_ context.Context, userID, name string) (Session, error) {
	id, err := newID()
	if err != nil {
		return Session{}, err
	}
	csrf, err := newID()
	if err != nil {
		return Session{}, err
	}
	now := m.now()
	sess := Session{ID: id, UserID: userID, Name: name, CSRFToken: csrf, CreatedAt: now, ExpiresAt: m.expireAt(now)}
	m.mu.Lock()
	m.sessions[id] = sess
	m.mu.Unlock()
	return sess, nil
}

func (m *Memory) Get(_ context.Context, id string) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	sess, ok := m.sessions[id]
	if !ok || !m.now().Before(sess.ExpiresAt) {
		delete(m.sessions, id)
		return Session{}, ErrNotFound
	}
	sess.ExpiresAt = m.expireAt(sess.CreatedAt)
	if !m.now().Before(sess.ExpiresAt) {
		delete(m.sessions, id)
		return Session{}, ErrNotFound
	}
	m.sessions[id] = sess
	return sess, nil
}

func (m *Memory) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	delete(m.sessions, id)
	m.mu.Unlock()
	return nil
}

func (m *Memory) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	now := m.now()
	for _, sess := range m.sessions {
		if now.Before(sess.ExpiresAt) {
			n++
		}
	}
	return n
}

var _ Store = (*Memory)(nil)
