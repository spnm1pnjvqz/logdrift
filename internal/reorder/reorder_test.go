package reorder_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/reorder"
	"github.com/user/logdrift/internal/runner"
)

func makeLineCh(lines []runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collect(ch <-chan runner.LogLine) []runner.LogLine {
	var out []runner.LogLine
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func fixedExtract(ts time.Time) func(string) (time.Time, bool) {
	return func(text string) (time.Time, bool) {
		return ts, true
	}
}

func TestNew_InvalidWindow_ReturnsError(t *testing.T) {
	_, err := reorder.New(0, fixedExtract(time.Now()))
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestNew_NilExtract_ReturnsError(t *testing.T) {
	_, err := reorder.New(50*time.Millisecond, nil)
	if err == nil {
		t.Fatal("expected error for nil extract")
	}
}

func TestNew_ValidConfig_NoError(t *testing.T) {
	_, err := reorder.New(50*time.Millisecond, fixedExtract(time.Now()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_SortsOutOfOrderLines(t *testing.T) {
	now := time.Now()
	t1 := now.Add(-2 * time.Second)
	t2 := now.Add(-1 * time.Second)
	t3 := now

	times := []time.Time{t3, t1, t2}
	idx := 0
	extract := func(text string) (time.Time, bool) {
		ts := times[idx%len(times)]
		idx++
		return ts, true
	}

	r, err := reorder.New(80*time.Millisecond, extract)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	src := makeLineCh([]runner.LogLine{
		{Service: "svc", Text: "third"},
		{Service: "svc", Text: "first"},
		{Service: "svc", Text: "second"},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	lines := collect(r.Apply(ctx, src))
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0].Text != "first" || lines[1].Text != "second" || lines[2].Text != "third" {
		t.Errorf("unexpected order: %v", lines)
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	r, _ := reorder.New(500*time.Millisecond, fixedExtract(time.Now()))

	blocking := make(chan runner.LogLine) // never closed
	ctx, cancel := context.WithCancel(context.Background())
	out := r.Apply(ctx, blocking)
	cancel()

	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for channel close")
	}
}
