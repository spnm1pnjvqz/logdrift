package rotate

import (
	"context"
	"os"
	"testing"
	"time"
)

func writeTempLog(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "rotate-*.log")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestWatcher_NoRotation_NoEvents(t *testing.T) {
	path := writeTempLog(t, "hello\n")
	w := New(map[string]string{"svc": path}, 50*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	ch := w.Watch(ctx)
	var events []Event
	for ev := range ch {
		events = append(events, ev)
	}
	// first tick initialises state; subsequent ticks should see no change
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
}

func TestWatcher_FileShrinks_EmitsEvent(t *testing.T) {
	path := writeTempLog(t, "aaabbbccc\n")
	w := New(map[string]string{"svc": path}, 40*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	ch := w.Watch(ctx)
	// Allow one poll to establish baseline.
	time.Sleep(60 * time.Millisecond)
	// Truncate the file to simulate rotation.
	if err := os.WriteFile(path, []byte("x\n"), 0o644); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	var got Event
	select {
	case ev, ok := <-ch:
		if !ok {
			t.Fatal("channel closed before event")
		}
		got = ev
	case <-time.After(400 * time.Millisecond):
		t.Fatal("timed out waiting for rotation event")
	}
	cancel()
	if got.Source != "svc" {
		t.Errorf("source = %q, want \"svc\"", got.Source)
	}
	if got.Path != path {
		t.Errorf("path = %q, want %q", got.Path, path)
	}
}

func TestWatcher_CancelClosesChannel(t *testing.T) {
	path := writeTempLog(t, "data\n")
	w := New(map[string]string{"svc": path}, 30*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	ch := w.Watch(ctx)
	cancel()
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected channel to be closed")
		}
	case <-time.After(300 * time.Millisecond):
		t.Error("channel not closed after cancel")
	}
}

func TestNew_DefaultInterval(t *testing.T) {
	w := New(map[string]string{}, 0)
	if w.pollInterval != DefaultPollInterval {
		t.Errorf("interval = %v, want %v", w.pollInterval, DefaultPollInterval)
	}
}
