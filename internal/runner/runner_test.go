package runner

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"
)

func echoArgs() (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/C", "echo hello"}
	}
	return "sh", []string{"-c", "echo hello"}
}

func TestRunner_Start_ReceivesLines(t *testing.T) {
	r := New()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shell, args := echoArgs()
	ch, err := r.Start(ctx, "svc", shell, args)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	var lines []Line
	for l := range ch {
		lines = append(lines, l)
	}

	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
	if lines[0].Service != "svc" {
		t.Errorf("service = %q, want %q", lines[0].Service, "svc")
	}
	if !strings.Contains(lines[0].Text, "hello") {
		t.Errorf("text = %q, want to contain 'hello'", lines[0].Text)
	}
}

func TestRunner_Start_ChannelClosedOnExit(t *testing.T) {
	r := New()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shell, args := echoArgs()
	ch, err := r.Start(ctx, "svc", shell, args)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	for range ch {
	}
	// If we reach here the channel was closed — pass.
}

func TestRunner_StopAll(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sleep not available on windows")
	}
	r := New()
	ctx := context.Background()

	_, err := r.Start(ctx, "svc", "sh", []string{"-c", "sleep 60"})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	r.StopAll()

	r.mu.Lock()
	defer r.mu.Unlock()
	for _, cmd := range r.cmds {
		if cmd.ProcessState == nil && cmd.Process != nil {
			// Give the OS a moment to reap.
			time.Sleep(100 * time.Millisecond)
		}
	}
}
