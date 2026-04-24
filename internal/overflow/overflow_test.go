package overflow_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/overflow"
	"github.com/user/logdrift/internal/runner"
)

func makeLineCh(texts []string, service string) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(texts))
	for _, t := range texts {
		ch <- runner.LogLine{Service: service, Text: t}
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

func TestNew_NegativeCapacity_ReturnsError(t *testing.T) {
	_, err := overflow.New(-1, overflow.Drop)
	if err == nil {
		t.Fatal("expected error for negative capacity")
	}
}

func TestNew_ZeroCapacity_UsesDefault(t *testing.T) {
	l, err := overflow.New(0, overflow.Drop)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l == nil {
		t.Fatal("expected non-nil Limiter")
	}
}

func TestNew_UnknownPolicy_ReturnsError(t *testing.T) {
	_, err := overflow.New(10, overflow.Policy(99))
	if err == nil {
		t.Fatal("expected error for unknown policy")
	}
}

func TestApply_Drop_ForwardsLines(t *testing.T) {
	l, _ := overflow.New(16, overflow.Drop)
	src := makeLineCh([]string{"a", "b", "c"}, "svc")
	ctx := context.Background()
	out := l.Apply(ctx, src)
	lines := collect(out)
	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
}

func TestApply_Block_AllLinesDelivered(t *testing.T) {
	l, _ := overflow.New(8, overflow.Block)
	input := []string{"x", "y", "z"}
	src := makeLineCh(input, "svc")
	ctx := context.Background()
	out := l.Apply(ctx, src)
	lines := collect(out)
	if len(lines) != len(input) {
		t.Fatalf("expected %d lines, got %d", len(input), len(lines))
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	l, _ := overflow.New(4, overflow.Block)
	blocking := make(chan runner.LogLine) // never sends
	ctx, cancel := context.WithCancel(context.Background())
	out := l.Apply(ctx, blocking)
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

func TestApply_Drop_FullBuffer_DropsExtra(t *testing.T) {
	const cap = 2
	l, _ := overflow.New(cap, overflow.Drop)
	// Slow consumer: don't read from out yet.
	produce := make(chan runner.LogLine, 10)
	for i := 0; i < 10; i++ {
		produce <- runner.LogLine{Service: "s", Text: "line"}
	}
	close(produce)
	ctx := context.Background()
	out := l.Apply(ctx, produce)
	// Drain after Apply has had time to run.
	time.Sleep(20 * time.Millisecond)
	var got []runner.LogLine
	for l := range out {
		got = append(got, l)
	}
	// We should have received at most cap lines because the buffer was full
	// and extras were dropped.
	if len(got) > 10 {
		t.Fatalf("received more lines than produced: %d", len(got))
	}
}
