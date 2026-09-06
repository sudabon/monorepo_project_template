package server

import (
	"testing"
	"time"
)

func TestShutdownTimeoutConfiguration(t *testing.T) {
	for _, tc := range []struct {
		value string
		valid bool
	}{{"", true}, {"125ms", true}, {"5s", true}, {"0s", false}, {"-1s", false}, {"wrong", false}} {
		t.Setenv("SHUTDOWN_TIMEOUT", tc.value)
		wait, err := ShutdownTimeout()
		if (err == nil) != tc.valid || (tc.valid && wait <= 0) {
			t.Fatalf("%q = %s, %v", tc.value, wait, err)
		}
		if tc.value == "125ms" && wait != 125*time.Millisecond {
			t.Fatalf("configured duration = %s", wait)
		}
	}
}
