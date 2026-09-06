package handler_test

import (
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sudabon/monorepo_project_template/apps/bff/internal/session"
)

func TestCurlLoginProtectedLogoutAndCSRF(t *testing.T) {
	if _, err := exec.LookPath("curl"); err != nil {
		t.Skip("curl not installed")
	}
	h, stub := newBFF(t, nil)
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	dir := t.TempDir()
	loginHdr, loginBody := filepath.Join(dir, "login.hdr"), filepath.Join(dir, "login.body")
	curl(t, "-sS", "-D", loginHdr, "-o", loginBody, "-X", "POST", srv.URL+"/auth/login",
		"-H", "Content-Type: application/json",
		"-d", `{"username":"demo","password":"secret"}`)
	sid := cookieValue(t, readFile(t, loginHdr), session.CookieName)
	csrf := jsonField(t, readFile(t, loginBody), "csrfToken")
	protected := curl(t, "-sS", "-o", "/dev/null", "-w", "%{http_code}",
		"-H", "Cookie: "+session.CookieName+"="+sid, srv.URL+"/api/items")
	if protected != "200" {
		t.Fatalf("protected = %s", protected)
	}
	_ = curl(t, "-sS", "-o", "/dev/null", "-X", "POST", srv.URL+"/auth/logout",
		"-H", "Cookie: "+session.CookieName+"="+sid,
		"-H", "X-CSRF-Token: "+csrf)
	after := curl(t, "-sS", "-o", "/dev/null", "-w", "%{http_code}",
		"-H", "Cookie: "+session.CookieName+"="+sid, srv.URL+"/api/items")
	if after != "401" {
		t.Fatalf("after logout = %s", after)
	}
	cookie, _ := loginOK(t, h)
	denied := curl(t, "-sS", "-o", "/dev/null", "-w", "%{http_code}",
		"-X", "POST", srv.URL+"/api/items",
		"-H", "Content-Type: application/json",
		"-H", "Cookie: "+session.CookieName+"="+cookie.Value,
		"-d", `{"name":"nope"}`)
	if denied != "403" {
		t.Fatalf("csrf missing = %s", denied)
	}
	if stub.calls < 1 {
		t.Fatal("expected at least the GET before logout")
	}
}

func curl(t *testing.T, args ...string) string {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "curl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("curl %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func cookieValue(t *testing.T, raw, name string) string {
	t.Helper()
	for line := range strings.SplitSeq(raw, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(strings.ToLower(line), "set-cookie:") {
			continue
		}
		_, body, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		parts := strings.Split(strings.TrimSpace(body), ";")
		key, value, ok := strings.Cut(parts[0], "=")
		if ok && strings.TrimSpace(key) == name {
			return strings.TrimSpace(value)
		}
	}
	t.Fatalf("missing %s in %s", name, raw)
	return ""
}

func jsonField(t *testing.T, raw, field string) string {
	t.Helper()
	needle := `"` + field + `":"`
	_, rest, ok := strings.Cut(raw, needle)
	if !ok {
		t.Fatalf("missing %s in %s", field, raw)
	}
	value, _, ok := strings.Cut(rest, `"`)
	if !ok {
		t.Fatalf("unterminated %s in %s", field, raw)
	}
	return value
}
