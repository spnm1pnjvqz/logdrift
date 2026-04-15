package snapshot_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/runner"
	"github.com/yourorg/logdrift/internal/snapshot"
)

func makeLogLineCh(lines []runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestCollector_CollectsAllLines(t *testing.T) {
	lines := []runner.LogLine{
		{Service: "svc-a", Text: "first"},
		{Service: "svc-b", Text: "second"},
		{Service: "svc-a", Text: "third"},
	}
	ch := makeLogLineCh(lines)

	c := snapshot.NewCollector()
	snap := c.Collect(context.Background(), ch)

	if len(snap.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(snap.Entries))
	}
	for i, e := range snap.Entries {
		if e.Service != lines[i].Service {
			t.Errorf("entry %d: service mismatch: want %s got %s", i, lines[i].Service, e.Service)
		}
		if e.Line != lines[i].Text {
			t.Errorf("entry %d: line mismatch: want %s got %s", i, lines[i].Text, e.Line)
		}
	}
}

func TestCollector_CancelStopsCollection(t *testing.T) {
	// Unbuffered channel — nothing will be sent.
	ch := make(chan runner.LogLine)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	c := snapshot.NewCollector()
	snap := c.Collect(ctx, ch)

	// Should return quickly with an empty snapshot.
	if len(snap.Entries) != 0 {
		t.Fatalf("expected 0 entries after cancel, got %d", len(snap.Entries))
	}
}

func TestCollector_SnapshotMethod_ReturnsSameInstance(t *testing.T) {
	lines := []runner.LogLine{{Service: "x", Text: "y"}}
	ch := makeLogLineCh(lines)

	c := snapshot.NewCollector()
	collected := c.Collect(context.Background(), ch)
	via := c.Snapshot()

	if collected != via {
		t.Error("Snapshot() should return the same instance as Collect()")
	}
}
