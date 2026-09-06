package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestContextIDReachesEveryLayerLog(t *testing.T) {
	var out bytes.Buffer
	log := New(&out)
	ctx := WithRequestID(context.Background(), "trace-123")
	log.InfoContext(ctx, "handler")
	log.With("layer", "usecase").ErrorContext(ctx, "failed")
	for _, line := range bytes.Split(bytes.TrimSpace(out.Bytes()), []byte("\n")) {
		var entry map[string]any
		if err := json.Unmarshal(line, &entry); err != nil {
			t.Fatal(err)
		}
		if entry["request_id"] != "trace-123" {
			t.Fatalf("log = %v", entry)
		}
	}
	if RequestID(ctx) != "trace-123" || RequestID(context.Background()) != "" {
		t.Fatal("context ID retrieval failed")
	}
	if !log.Enabled(ctx, slog.LevelInfo) {
		t.Fatal("info logs disabled")
	}
}
