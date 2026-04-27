package debounce

import (
	"context"
	"testing"
	"time"

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

func TestApply_UniqueLines_AllForwarded(t *testing.T) {
	d, _ := New(200 * time.Millisecond)
	src := makeLineCh([]runner.LogLine{
		line("svc", "alpha"),
		line("svc", "beta"),
		line("svc", "gamma"),
	})
	out := Apply(context.Background(), d, src)
	got := collect(out)
	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(got))
	}
}

func TestApply_DuplicatesSuppressed(t *testing.T) {
	d, _ := New(200 * time.Millisecond)
	src := makeLineCh([]runner.LogLine{
		line("svc", "hello"),
		line("svc", "hello"),
		line("svc", "hello"),
	})
	out := Apply(context.Background(), d, src)
	got := collect(out)
	if len(got) != 1 {
		t.Fatalf("expected 1 line, got %d", len(got))
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	d, _ := New(200 * time.Millisecond)
	ch := make(chan runner.LogLine)
	ctx, cancel := context.WithCancel(context.Background())
	out := Apply(ctx, d, ch)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Error("expected channel to be closed after cancel")
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("timed out waiting for channel close")
	}
}

func TestApply_ClosesWhenInputClosed(t *testing.T) {
	d, _ := New(200 * time.Millisecond)
	src := makeLineCh(nil)
	out := Apply(context.Background(), d, src)
	select {
	case _, ok := <-out:
		if ok {
			t.Error("expected closed channel")
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("timed out")
	}
}
