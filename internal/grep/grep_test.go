package grep_test

import (
	"context"
	"testing"
	"time"

	"github.com/celrenheit/logdrift/internal/grep"
	"github.com/celrenheit/logdrift/internal/runner"
)

func makeLineCh(lines []string) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- runner.LogLine{Service: "svc", Text: l}
	}
	close(ch)
	return ch
}

func TestNew_NoPatterns_ReturnsError(t *testing.T) {
	_, err := grep.New(nil, false)
	if err == nil {
		t.Fatal("expected error for empty patterns")
	}
}

func TestNew_InvalidPattern_ReturnsError(t *testing.T) {
	_, err := grep.New([]string{"[invalid"}, false)
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestMatch_IncludeMode(t *testing.T) {
	g, err := grep.New([]string{"error", "warn"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !g.Match("an error occurred") {
		t.Error("expected match for 'error'")
	}
	if g.Match("all good") {
		t.Error("expected no match for unrelated line")
	}
}

func TestMatch_InvertMode(t *testing.T) {
	g, err := grep.New([]string{"debug"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if g.Match("debug: verbose") {
		t.Error("inverted: debug line should not match")
	}
	if !g.Match("info: started") {
		t.Error("inverted: non-debug line should match")
	}
}

func TestApply_FiltersLines(t *testing.T) {
	g, _ := grep.New([]string{"error"}, false)
	in := makeLineCh([]string{"ok line", "error: boom", "another ok", "error: again"})
	ctx := context.Background()
	out := g.Apply(ctx, in)

	var got []string
	for l := range out {
		got = append(got, l.Text)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(got), got)
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	g, _ := grep.New([]string{".*"}, false)
	block := make(chan runner.LogLine)
	ctx, cancel := context.WithCancel(context.Background())
	out := g.Apply(ctx, block)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Error("expected channel to be closed")
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for channel close")
	}
}
