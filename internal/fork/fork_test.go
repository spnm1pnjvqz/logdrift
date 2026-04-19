package fork_test

import (
	"context"
	"testing"
	"time"

	"github.com/logdrift/logdrift/internal/fork"
	"github.com/logdrift/logdrift/internal/runner"
)

func makeSrc(lines ...string) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- runner.LogLine{Service: "svc", Text: l}
	}
	close(ch)
	return ch
}

func collect(ch <-chan runner.LogLine) []string {
	var out []string
	for l := range ch {
		out = append(out, l.Text)
	}
	return out
}

func TestFork_EachOutputReceivesAllLines(t *testing.T) {
	src := makeSrc("a", "b", "c")
	outs := fork.Fork(context.Background(), src, 3, 8)
	if len(outs) != 3 {
		t.Fatalf("expected 3 outputs, got %d", len(outs))
	}
	for i, ch := range outs {
		got := collect(ch)
		if len(got) != 3 {
			t.Errorf("output %d: expected 3 lines, got %d", i, len(got))
		}
	}
}

func TestFork_OutputsClosedWhenSourceClosed(t *testing.T) {
	src := makeSrc("x")
	outs := fork.Fork(context.Background(), src, 2, 8)
	for i, ch := range outs {
		select {
		case _, ok := <-ch:
			_ = ok
		case <-time.After(time.Second):
			t.Errorf("output %d never closed", i)
		}
		// drain to close
		for range ch {
		}
	}
}

func TestFork_CancelStopsOutput(t *testing.T) {
	blocking := make(chan runner.LogLine)
	ctx, cancel := context.WithCancel(context.Background())
	outs := fork.Fork(ctx, blocking, 2, 0)
	cancel()
	for i, ch := range outs {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Errorf("output %d did not close after cancel", i)
		}
	}
}
