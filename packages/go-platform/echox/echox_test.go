package echox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sudabon/monorepo_project_template/packages/go-platform/logging"
)

func TestTraceRequestAssignsAndPropagatesID(t *testing.T) {
	var capturedReq, capturedResp string
	e := echo.New()
	e.Use(TraceRequest)
	e.GET("/", func(c echo.Context) error {
		capturedReq = c.Request().Header.Get(logging.RequestIDHeader)
		capturedResp = c.Response().Header().Get(logging.RequestIDHeader)
		return c.NoContent(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	generated := rec.Header().Get(logging.RequestIDHeader)
	if generated == "" || capturedReq != generated || capturedResp != generated {
		t.Fatalf("generated id req=%q resp=%q header=%q", capturedReq, capturedResp, generated)
	}

	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(logging.RequestIDHeader, "upstream-123")
	e.ServeHTTP(rec, req)
	if rec.Header().Get(logging.RequestIDHeader) != "upstream-123" || capturedReq != "upstream-123" || capturedResp != "upstream-123" {
		t.Fatalf("propagated req=%q resp=%q header=%q", capturedReq, capturedResp, rec.Header().Get(logging.RequestIDHeader))
	}
}

func TestTraceRequestHidesPanicFromBody(t *testing.T) {
	var captured error
	e := echo.New()
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		captured = err
		WriteError(c, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
	}
	e.Use(TraceRequest)
	e.GET("/panic", func(echo.Context) error { panic("secret-panic") })

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/panic", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("secret-panic")) || bytes.Contains(rec.Body.Bytes(), []byte("goroutine")) {
		t.Fatalf("body leaked internals: %s", rec.Body)
	}
	if captured == nil || !strings.Contains(captured.Error(), "request panic:") || !strings.Contains(captured.Error(), "secret-panic") || !strings.Contains(captured.Error(), "goroutine") {
		t.Fatalf("panic error lost the stack: %v", captured)
	}
}

func TestHealthRoutesShallowStaysUpWhenPingFails(t *testing.T) {
	e := echo.New()
	HealthRoutes(e, func(context.Context) error { return nil })
	assertHealth(t, e, "/health/shallow", http.StatusOK, "healthy")
	assertHealth(t, e, "/health/deep", http.StatusOK, "healthy")

	e = echo.New()
	HealthRoutes(e, func(context.Context) error { return errors.New("db down") })
	assertHealth(t, e, "/health/shallow", http.StatusOK, "healthy")
	assertHealth(t, e, "/health/deep", http.StatusServiceUnavailable, "unhealthy")
}

func TestWriteErrorSkipsBodyWhenCommitted(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(logging.New(&logs))
	t.Cleanup(func() { slog.SetDefault(previous) })

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		if err := c.String(http.StatusOK, "already sent"); err != nil {
			return err
		}
		WriteError(c, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
		return nil
	})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "already sent" {
		t.Fatalf("committed write changed the response: %d %s", rec.Code, rec.Body)
	}
	if !bytes.Contains(logs.Bytes(), []byte("response failed after headers were sent")) {
		t.Fatalf("missing committed log: %s", logs.String())
	}
}

func assertHealth(t *testing.T, e *echo.Echo, path string, status int, want string) {
	t.Helper()
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != status {
		t.Fatalf("%s status = %d %s", path, rec.Code, rec.Body)
	}
	var body struct {
		Status       string            `json:"status"`
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != want {
		t.Fatalf("%s status field = %q", path, body.Status)
	}
	if path == "/health/deep" && body.Dependencies["database"] != want {
		t.Fatalf("deep dependencies = %v", body.Dependencies)
	}
}
