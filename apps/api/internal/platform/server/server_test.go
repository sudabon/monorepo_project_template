package server

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
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

func TestSIGTERMDrainsAndEnforcesDeadline(t *testing.T) {
	for _, tc := range []struct {
		name, timeout string
		complete      bool
	}{{"drain", "3s", true}, {"deadline", "100ms", false}} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestServerProcess$")
			cmd.Env = append(os.Environ(), "API_SERVER_TEST_PROCESS=1", "SHUTDOWN_TIMEOUT="+tc.timeout)
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			var logs bytes.Buffer
			cmd.Stderr = &logs
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if cmd.ProcessState == nil {
					_ = cmd.Process.Kill()
					_ = cmd.Wait()
				}
			})
			address, err := bufio.NewReader(stdout).ReadString('\n')
			if err != nil {
				t.Fatal(err)
			}
			client := &http.Client{Timeout: 5 * time.Second}
			response, err := client.Get("http://" + strings.TrimSpace(address) + "/slow")
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			reader := bufio.NewReader(response.Body)
			started, err := reader.ReadString('\n')
			if err != nil || started != "started\n" {
				t.Fatalf("start = %q, %v", started, err)
			}
			start := time.Now()
			if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
				t.Fatal(err)
			}
			body, readErr := io.ReadAll(reader)
			waitErr := cmd.Wait()
			elapsed := time.Since(start)
			if waitErr != nil {
				t.Fatalf("process = %v, logs=%s", waitErr, logs.String())
			}
			if tc.complete {
				if readErr != nil || string(body) != "completed\n" || !strings.Contains(logs.String(), "shutdown completed") {
					t.Fatalf("drain: body=%q err=%v logs=%s", body, readErr, logs.String())
				}
			} else {
				if readErr == nil || strings.Contains(string(body), "completed") || !strings.Contains(logs.String(), "shutdown deadline reached") {
					t.Fatalf("deadline: body=%q err=%v logs=%s", body, readErr, logs.String())
				}
				if elapsed > 2*time.Second {
					t.Fatalf("deadline not enforced: %s", elapsed)
				}
			}
		})
	}
}

// Runs the production signal/shutdown path in a subprocess so SIGTERM is real.
func TestServerProcess(t *testing.T) {
	if os.Getenv("API_SERVER_TEST_PROCESS") != "1" {
		return
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		os.Exit(2)
	}
	timeout, err := ShutdownTimeout()
	if err != nil {
		os.Exit(2)
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "started\n")
		w.(http.Flusher).Flush()
		select {
		case <-time.After(400 * time.Millisecond):
			_, _ = io.WriteString(w, "completed\n")
		case <-r.Context().Done():
		}
	})
	// The listener address is printed by the handler setup before the parent makes
	// its request; Serve installs signal handling before accepting that request.
	fmt.Println(listener.Addr().String())
	if err := Serve(listener, handler, timeout); err != nil {
		os.Exit(2)
	}
	os.Exit(0)
}
