package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/sudabon/monorepo_project_template/apps/bff/internal/handler"
	"github.com/sudabon/monorepo_project_template/apps/bff/internal/identity"
	"github.com/sudabon/monorepo_project_template/apps/bff/internal/session"
)

// brokenStore stands in for a database outage rather than a missing session.
type brokenStore struct{ session.Store }

func (brokenStore) Get(context.Context, string) (session.Session, error) {
	return session.Session{}, errors.New("dial tcp: connection refused")
}

func TestStoreOutageIsServerErrorNotUnauthenticated(t *testing.T) {
	backend := &backendStub{}
	srv := httptest.NewServer(backend)
	t.Cleanup(srv.Close)
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	h := handler.New(handler.Deps{
		Store:        brokenStore{session.NewMemory()},
		Users:        identity.Static{Username: "demo", Password: "secret", User: identity.User{ID: "user-1", Name: "Demo"}},
		Backend:      u,
		CookieSecure: true,
		Ping:         func(context.Context) error { return nil },
	})
	rec := request(t, h, http.MethodGet, "/api/items", "", []*http.Cookie{{Name: session.CookieName, Value: "still-valid-in-the-store"}}, nil)
	// 401 here would sign every user out for as long as the store is down.
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("store outage = %d %s", rec.Code, rec.Body)
	}
	if strings.Contains(rec.Body.String(), "connection refused") {
		t.Fatalf("response leaked the store error: %s", rec.Body)
	}
	if backend.calls != 0 {
		t.Fatal("backend must not be called without a session")
	}
}

func TestBrowserCredentialsAreNotForwardedToBackend(t *testing.T) {
	h, stub := newBFF(t, nil)
	cookie, _ := loginOK(t, h)
	hdr := http.Header{"Authorization": []string{"Bearer client-supplied"}}
	rec := request(t, h, http.MethodGet, "/api/items", "", []*http.Cookie{cookie}, hdr)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d %s", rec.Code, rec.Body)
	}
	got, _ := stub.last()
	if got.Get("Cookie") != "" {
		t.Fatalf("session cookie reached the backend: %q", got.Get("Cookie"))
	}
	if got.Get("Authorization") != "" {
		t.Fatalf("client Authorization reached the backend: %q", got.Get("Authorization"))
	}
	if got.Get(identity.UserIDHeader) != "user-1" {
		t.Fatalf("identity header = %q", got.Get(identity.UserIDHeader))
	}
}

func TestLoginBodyIsCapped(t *testing.T) {
	h, _ := newBFF(t, nil)
	oversized := `{"username":"demo","password":"` + strings.Repeat("a", 64<<10) + `"}`
	rec := request(t, h, http.MethodPost, "/auth/login", oversized, nil, nil)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized login = %d %s", rec.Code, rec.Body)
	}
	if cookieNamed(rec, session.CookieName) != nil {
		t.Fatal("oversized login must not create a session")
	}
}
