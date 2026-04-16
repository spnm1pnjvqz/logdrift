package colorize

import (
	"context"
	"strings"
	"testing"

	"github.com/iamcathal/logdrift/internal/runner"
)

func makeLineCh(lines []runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestApply_WrapsText(t *testing.T) {
	c := New()
	lines := []runner.LogLine{
		{Service: "api", Text: "started"},
		{Service: "db", Text: "connected"},
	}
	in := makeLineCh(lines)
	out := Apply(context.Background(), c, in)
	var got []runner.LogLine
	for l := range out {
		got = append(got, l)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
	for _, l := range got {
		if !strings.Contains(l.Text, "\x1b[") {
			t.Errorf("line %q missing ANSI escape", l.Text)
		}
	}
}

func TestApply_PreservesService(t *testing.T) {
	c := New()
	in := makeLineCh([]runner.LogLine{{Service: "worker", Text: "job done"}})
	out := Apply(context.Background(), c, in)
	l := <-out
	if l.Service != "worker" {
		t.Fatalf("expected service 'worker', got %q", l.Service)
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	c := New()
	blocking := make(chan runner.LogLine)
	ctx, cancel := context.WithCancel(context.Background())
	out := Apply(ctx, c, blocking)
	cancel()
	_, ok := <-out
	if ok {
		t.Fatal("expected channel to be closed after cancel")
	}
}

func TestApply_ClosesWhenInputClosed(t *testing.T) {
	c := New()
	in := makeLineCh(nil)
	out := Apply(context.Background(), c, in)
	_, ok := <-out
	if ok {
		t.Fatal("expected closed channel")
	}
}
