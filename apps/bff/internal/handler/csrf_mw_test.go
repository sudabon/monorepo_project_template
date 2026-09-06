package handler_test

import (
	"os"
	"strings"
	"testing"
)

func TestCSRFMiddlewareIsGlobalNotPerRouteOptIn(t *testing.T) {
	src, err := os.ReadFile("router.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	if !strings.Contains(text, "e.Use(protectCSRF)") {
		t.Fatal("CSRF must be registered with e.Use so every state-changing method is covered")
	}
	if strings.Contains(text, "protectCSRF)") && strings.Contains(text, "e.POST") {
		for _, line := range strings.Split(text, "\n") {
			trim := strings.TrimSpace(line)
			if strings.Contains(trim, "protectCSRF") && (strings.HasPrefix(trim, "e.POST") || strings.HasPrefix(trim, "e.PUT") || strings.HasPrefix(trim, "e.PATCH") || strings.HasPrefix(trim, "e.DELETE") || strings.Contains(trim, "Group(")) {
				t.Fatalf("CSRF looks like per-route opt-in: %s", trim)
			}
		}
	}
}
