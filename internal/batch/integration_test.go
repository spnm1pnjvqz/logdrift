package batch

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/runner"
)

// TestApply_IntervalFlushesPartialBatch verifies that a partial batch is
// emitted when the flush interval elapses before the batch is full.
func TestApply_IntervalFlushesPartialBatch(t *testing.T) {
	b, err := New(100, 60*time.Millisecond)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	src := make(chan runner.LogLine, 2)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out := b.Apply(ctx, src)

	// Send 2 lines — far fewer than the batch size of 100.
	src <- runner.LogLine{Service: "api", Text: "line1"}
	src <- runner.LogLine{Service: "api", Text: "line2"}

	// Wait for the interval to fire and flush.
	select {
	case batch := <-out:
		if len(batch) != 2 {
			t.Fatalf("expected 2 lines in interval-flushed batch, got %d", len(batch))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for interval flush")
	}
}
