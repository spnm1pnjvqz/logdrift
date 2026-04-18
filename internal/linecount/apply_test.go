package linecount

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/runner"
)

func makeLinesForApply(lines []runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestApply_CountsAndForwardsLines(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	input := []runner.LogLine{
		{Service: "svc-a", Text: "hello"},
		{Service: "svc-a", Text: "world"},
		{Service: "svc-b", Text: "foo"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	out := c.Apply(ctx, makeLinesForApply(input))

	var got []runner.LogLine
	for l := range out {
		got = append(got, l)
	}

	if len(got) != len(input) {
		t.Fatalf("expected %d lines, got %d", len(input), len(got))
	}
	if c.Get("svc-a") != 2 {
		t.Errorf("svc-a count: want 2, got %d", c.Get("svc-a"))
	}
	if c.Get("svc-b") != 1 {
		t.Errorf("svc-b count: want 1, got %d", c.Get("svc-b"))
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	c, _ := New()
	ch := make(chan runner.LogLine) // never sends
	ctx, cancel := context.WithCancel(context.Background())
	out := c.Apply(ctx, ch)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("channel not closed after cancel")
	}
}
