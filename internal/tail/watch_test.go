package tail_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/tail"
)

func TestWatch_MissingFile_ReturnsError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := tail.Watch(ctx, tail.WatchConfig{Path: "/nonexistent/file.log"})
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestWatch_ExistingFile_EmitsInitialResult(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "watch-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	results, err := tail.Watch(ctx, tail.WatchConfig{
		Path:     f.Name(),
		Interval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case r, ok := <-results:
		if !ok {
			t.Fatal("results channel closed before emitting initial result")
		}
		if r.Path != f.Name() {
			t.Errorf("got path %q, want %q", r.Path, f.Name())
		}
		if r.Lines == nil {
			t.Error("Lines channel is nil")
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for initial WatchResult")
	}
}

func TestWatch_Cancel_ClosesResultsChannel(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "watch-cancel-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	ctx, cancel := context.WithCancel(context.Background())

	results, err := tail.Watch(ctx, tail.WatchConfig{
		Path:     f.Name(),
		Interval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Drain the initial result.
	select {
	case <-results:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for initial result")
	}

	cancel()

	select {
	case _, ok := <-results:
		if ok {
			// Drain any extra results from the restart loop.
			for range results {
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("results channel not closed after cancel")
	}
}
