package handler_test

import (
	"encoding/json/v2"
	"net/http"
	"strings"
	"testing"

	"github.com/sudabon/monorepo_project_template/apps/bff/internal/session"
	"github.com/sudabon/monorepo_project_template/packages/go-platform/logging"
)

func TestLoginSetsHttpOnlySecureLaxCookie(t *testing.T) {
	h, _ := newBFF(t, nil)
	rec := request(t, h, http.MethodPost, "/auth/login", `{"username":"demo","password":"secret"}`, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d %s", rec.Code, rec.Body)
	}
	c := cookieNamed(rec, session.CookieName)
	if c == nil {
		t.Fatal("missing session cookie")
	}
	if !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteLaxMode || c.Path != "/" || c.Value == "" {
		t.Fatalf("cookie = %+v", c)
	}
}

func TestLoginRejectsWrongCredentialsWithoutCookie(t *testing.T) {
	h, _ := newBFF(t, nil)
	rec := request(t, h, http.MethodPost, "/auth/login", `{"username":"demo","password":"wrong"}`, nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d %s", rec.Code, rec.Body)
	}
	if cookieNamed(rec, session.CookieName) != nil {
		t.Fatal("session cookie must not be set")
	}
}

func TestLoginResponseDoesNotLeakSessionID(t *testing.T) {
	h, _ := newBFF(t, nil)
	rec := request(t, h, http.MethodPost, "/auth/login", `{"username":"demo","password":"secret"}`, nil, nil)
	c := cookieNamed(rec, session.CookieName)
	if c == nil {
		t.Fatal("missing session cookie")
	}
	if strings.Contains(rec.Body.String(), c.Value) {
		t.Fatalf("body leaked session id: %s", rec.Body)
	}
	for key, values := range rec.Header() {
		if strings.EqualFold(key, "Set-Cookie") {
			continue
		}
		for _, value := range values {
			if strings.Contains(value, c.Value) {
				t.Fatalf("header %s leaked session id: %q", key, value)
			}
		}
	}
}

func TestSessionStatusAuthenticatedAndAnonymous(t *testing.T) {
	h, _ := newBFF(t, nil)
	anon := request(t, h, http.MethodGet, "/auth/session", "", nil, nil)
	if anon.Code != http.StatusOK {
		t.Fatalf("anon status = %d", anon.Code)
	}
	var anonView struct {
		Authenticated bool `json:"authenticated"`
	}
	if err := json.Unmarshal(anon.Body.Bytes(), &anonView); err != nil {
		t.Fatal(err)
	}
	if anonView.Authenticated {
		t.Fatal("expected unauthenticated")
	}
	cookie, csrf := loginOK(t, h)
	authed := request(t, h, http.MethodGet, "/auth/session", "", []*http.Cookie{cookie}, nil)
	if authed.Code != http.StatusOK {
		t.Fatalf("authed status = %d %s", authed.Code, authed.Body)
	}
	var view struct {
		Authenticated bool `json:"authenticated"`
		User          struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"user"`
		CSRFToken string `json:"csrfToken"`
	}
	if err := json.Unmarshal(authed.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	if !view.Authenticated || view.User.ID != "user-1" || view.User.Name != "Demo" || view.CSRFToken != csrf {
		t.Fatalf("view = %+v body=%s", view, authed.Body)
	}
}

func TestLogoutRevokesSessionCookieReuse(t *testing.T) {
	h, stub := newBFF(t, nil)
	cookie, csrf := loginOK(t, h)
	got := request(t, h, http.MethodGet, "/api/items", "", []*http.Cookie{cookie}, nil)
	if got.Code != http.StatusOK {
		t.Fatalf("before logout = %d %s", got.Code, got.Body)
	}
	out := request(t, h, http.MethodPost, "/auth/logout", "", []*http.Cookie{cookie}, http.Header{handlerCSRF: []string{csrf}})
	if out.Code != http.StatusNoContent {
		t.Fatalf("logout = %d %s", out.Code, out.Body)
	}
	reused := request(t, h, http.MethodGet, "/api/items", "", []*http.Cookie{cookie}, nil)
	if reused.Code != http.StatusUnauthorized {
		t.Fatalf("reused cookie = %d %s", reused.Code, reused.Body)
	}
	if stub.calls < 1 {
		t.Fatal("protected resource was not fetched before logout")
	}
}

const handlerCSRF = "X-CSRF-Token"

func TestUnauthenticatedProtectedResourceIs401NotRedirect(t *testing.T) {
	h, stub := newBFF(t, nil)
	rec := request(t, h, http.MethodGet, "/api/items", "", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d %s", rec.Code, rec.Body)
	}
	if rec.Header().Get("Location") != "" {
		t.Fatalf("redirected to %s", rec.Header().Get("Location"))
	}
	if stub.calls != 0 {
		t.Fatal("backend must not be called")
	}
}

func TestInvalidSessionCookieIs401(t *testing.T) {
	h, _ := newBFF(t, nil)
	cookie, _ := loginOK(t, h)
	tampered := *cookie
	tampered.Value = cookie.Value[:len(cookie.Value)-2] + "ff"
	for _, c := range []*http.Cookie{&tampered, {Name: session.CookieName, Value: "revoked-or-unknown"}} {
		rec := request(t, h, http.MethodGet, "/api/items", "", []*http.Cookie{c}, nil)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d %s cookie=%s", rec.Code, rec.Body, c.Value)
		}
	}
	// Revoked cookie after logout is covered separately; expired is store-level.
}

func TestCSRFRequiredAndBoundToSession(t *testing.T) {
	h, stub := newBFF(t, nil)
	cookie, csrf := loginOK(t, h)
	missing := request(t, h, http.MethodPost, "/api/items", `{"name":"x"}`, []*http.Cookie{cookie}, nil)
	if missing.Code != http.StatusForbidden {
		t.Fatalf("missing csrf = %d %s", missing.Code, missing.Body)
	}
	if stub.calls != 0 {
		t.Fatal("state changed without csrf")
	}
	wrong := request(t, h, http.MethodPost, "/api/items", `{"name":"x"}`, []*http.Cookie{cookie}, http.Header{handlerCSRF: []string{"not-the-session-token"}})
	if wrong.Code != http.StatusForbidden {
		t.Fatalf("wrong csrf = %d %s", wrong.Code, wrong.Body)
	}
	ok := request(t, h, http.MethodPost, "/api/items", `{"name":"x"}`, []*http.Cookie{cookie}, http.Header{handlerCSRF: []string{csrf}})
	if ok.Code != http.StatusOK {
		t.Fatalf("valid csrf = %d %s", ok.Code, ok.Body)
	}
	if stub.calls != 1 {
		t.Fatalf("backend calls = %d", stub.calls)
	}
}

func TestForwardedIdentityAndRequestID(t *testing.T) {
	h, stub := newBFF(t, nil)
	cookie, _ := loginOK(t, h)
	hdr := http.Header{
		"X-User-ID":             []string{"spoofed"},
		logging.RequestIDHeader: []string{"client-trace"},
		"X-User-Id":             []string{"also-spoofed"},
	}
	rec := request(t, h, http.MethodGet, "/api/items", "", []*http.Cookie{cookie}, hdr)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d %s", rec.Code, rec.Body)
	}
	got, path := stub.last()
	if path != "/api/items" {
		t.Fatalf("path = %s", path)
	}
	if got.Get("X-User-ID") != "user-1" {
		t.Fatalf("user header = %q", got.Get("X-User-ID"))
	}
	if values := got.Values("X-User-ID"); len(values) != 1 || values[0] != "user-1" {
		t.Fatalf("user header values = %v", values)
	}
	if got.Get(logging.RequestIDHeader) != "client-trace" {
		t.Fatalf("request id = %q", got.Get(logging.RequestIDHeader))
	}
	if rec.Header().Get(logging.RequestIDHeader) != "client-trace" {
		t.Fatalf("response request id = %q", rec.Header().Get(logging.RequestIDHeader))
	}
}
