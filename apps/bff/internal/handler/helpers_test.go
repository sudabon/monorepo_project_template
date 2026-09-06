package handler_test

import (
	"context"
	"encoding/json/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/sudabon/monorepo_project_template/apps/bff/internal/handler"
	"github.com/sudabon/monorepo_project_template/apps/bff/internal/identity"
	"github.com/sudabon/monorepo_project_template/apps/bff/internal/session"
)

type backendStub struct {
	mu      sync.Mutex
	calls   int
	headers []http.Header
	paths   []string
	methods []string
	status  int
	body    string
}

func (s *backendStub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.calls++
	s.headers = append(s.headers, r.Header.Clone())
	s.paths = append(s.paths, r.URL.Path)
	s.methods = append(s.methods, r.Method)
	s.mu.Unlock()
	status := s.status
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_, _ = io.WriteString(w, s.body)
}

func (s *backendStub) last() (http.Header, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.headers) == 0 {
		return nil, ""
	}
	return s.headers[len(s.headers)-1], s.paths[len(s.paths)-1]
}

func newBFF(t *testing.T, backend http.Handler) (http.Handler, *backendStub) {
	t.Helper()
	stub, _ := backend.(*backendStub)
	if backend == nil {
		stub = &backendStub{}
		backend = stub
	}
	if stub == nil {
		stub = &backendStub{}
	}
	srv := httptest.NewServer(backend)
	t.Cleanup(srv.Close)
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	bff := handler.New(handler.Deps{
		Store:        session.NewMemory(),
		Users:        identity.Static{Username: "demo", Password: "secret", User: identity.User{ID: "user-1", Name: "Demo"}},
		Backend:      u,
		CookieSecure: true,
		Ping:         func(context.Context) error { return nil },
	})
	return bff, stub
}

func request(t *testing.T, h http.Handler, method, path, body string, cookies []*http.Cookie, hdr http.Header) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for k, vs := range hdr {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func loginOK(t *testing.T, h http.Handler) (*http.Cookie, string) {
	t.Helper()
	rec := request(t, h, http.MethodPost, "/auth/login", `{"username":"demo","password":"secret"}`, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("login = %d %s", rec.Code, rec.Body)
	}
	cookie := cookieNamed(rec, session.CookieName)
	if cookie == nil {
		t.Fatal("missing session cookie")
	}
	var view struct {
		CSRFToken string `json:"csrfToken"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	if view.CSRFToken == "" {
		t.Fatal("missing csrf token")
	}
	return cookie, view.CSRFToken
}

func cookieNamed(rec *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}
