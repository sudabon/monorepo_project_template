package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sudabon/monorepo_project_template/apps/api/internal/handler"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/platform/logging"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/repository"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/testdb"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/usecase"
)

func TestHealthAndErrorTracingWithUnavailableDB(t *testing.T) {
	db := testdb.Open(t)
	api := handler.New(usecase.NewItems(repository.NewItems(db)), db.PingContext)
	request(t, api, "GET", "/health/shallow", "", 200)
	request(t, api, "GET", "/health/deep", "", 200)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	// Closed pool represents loss of DB connectivity without mocking a healthy ping.
	request(t, api, "GET", "/health/shallow", "", 200)
	deep := request(t, api, "GET", "/health/deep", "", 503)
	var health struct {
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal(deep.Body.Bytes(), &health); err != nil {
		t.Fatal(err)
	}
	if health.Dependencies["database"] != "unhealthy" {
		t.Fatalf("deep = %s", deep.Body)
	}
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(logging.New(&logs))
	t.Cleanup(func() { slog.SetDefault(previous) })
	for _, id := range []string{"", "upstream-123"} {
		logs.Reset()
		req := httptest.NewRequest("GET", "/api/items", nil)
		req.Header.Set("X-Request-ID", id)
		rec := httptest.NewRecorder()
		api.ServeHTTP(rec, req)
		if rec.Code != 500 {
			t.Fatalf("status = %d", rec.Code)
		}
		responseID := rec.Header().Get("X-Request-ID")
		if responseID == "" || (id != "" && id != responseID) {
			t.Fatalf("ID = %q", responseID)
		}
		var body map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		if body["code"] != "internal_error" || bytes.Contains(rec.Body.Bytes(), []byte("database")) {
			t.Fatalf("error = %s", rec.Body)
		}
		errorLogged := false
		for _, line := range bytes.Split(bytes.TrimSpace(logs.Bytes()), []byte("\n")) {
			var entry map[string]any
			if err := json.Unmarshal(line, &entry); err != nil {
				t.Fatal(err)
			}
			if entry["request_id"] != responseID {
				t.Fatalf("log trace = %v", entry)
			}
			if entry["level"] == "ERROR" {
				errorLogged = true
			}
		}
		if !errorLogged {
			t.Fatalf("missing error log: %s", logs.String())
		}
	}
}

func TestShallowDoesNotCallDependencies(t *testing.T) {
	api := handler.New(nil, func(context.Context) error { t.Fatal("shallow pinged DB"); return nil })
	request(t, api, "GET", "/health/shallow", "", http.StatusOK)
}

type failingWriter struct{ *httptest.ResponseRecorder }

func (w failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("connection closed while writing")
}

func TestResponseWriteFailureIsLoggedWithRequestID(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(logging.New(&logs))
	t.Cleanup(func() { slog.SetDefault(previous) })
	api := handler.New(nil, nil)
	req := httptest.NewRequest("GET", "/health/shallow", nil)
	req.Header.Set("X-Request-ID", "failed-write")
	api.ServeHTTP(failingWriter{httptest.NewRecorder()}, req)
	found := false
	for _, line := range bytes.Split(bytes.TrimSpace(logs.Bytes()), []byte("\n")) {
		var entry map[string]any
		if err := json.Unmarshal(line, &entry); err != nil {
			t.Fatal(err)
		}
		if entry["level"] == "ERROR" && entry["request_id"] == "failed-write" {
			found = true
		}
	}
	if !found {
		t.Fatalf("write failure missing from logs: %s", logs.String())
	}
}
