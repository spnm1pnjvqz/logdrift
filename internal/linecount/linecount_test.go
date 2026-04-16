package linecount_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/linecount"
	"github.com/user/logdrift/internal/runner"
)

func TestAdd_EmptyService_ReturnsError(t *testing.T) {
	c := linecount.New()
	if err := c.Add(""); err == nil {
		t.Fatal("expected error for empty service")
	}
}

func TestAdd_IncrementsCount(t *testing.T) {
	c := linecount.New()
	_ = c.Add("svc-a")
	_ = c.Add("svc-a")
	_ = c.Add("svc-b")
	if got := c.Get("svc-a"); got != 2 {
		t.Fatalf("svc-a: want 2, got %d", got)
	}
	if got := c.Get("svc-b"); got != 1 {
		t.Fatalf("svc-b: want 1, got %d", got)
	}
}

func TestGet_UnknownService_ReturnsZero(t *testing.T) {
	c := linecount.New()
	if got := c.Get("missing"); got != 0 {
		t.Fatalf("want 0, got %d", got)
	}
}

func TestSnapshot_ReturnsCopy(t *testing.T) {
	c := linecount.New()
	_ = c.Add("x")
	snap := c.Snapshot()
	snap["x"] = 999
	if c.Get("x") != 1 {
		t.Fatal("snapshot mutation affected counter")
	}
}

func makeLineCh(lines []runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestApply_CountsAndForwards(t *testing.T) {
	c := linecount.New()
	lines := []runner.LogLine{
		{Service: "alpha", Text: "one"},
		{Service: "alpha", Text: "two"},
		{Service: "beta", Text: "three"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	out := c.Apply(ctx, makeLineCh(lines))
	var got []runner.LogLine
	for l := range out {
		got = append(got, l)
	}
	if len(got) != 3 {
		t.Fatalf("want 3 lines forwarded, got %d", len(got))
	}
	if c.Get("alpha") != 2 {
		t.Fatalf("alpha: want 2, got %d", c.Get("alpha"))
	}
	if c.Get("beta") != 1 {
		t.Fatalf("beta: want 1, got %d", c.Get("beta"))
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	c := linecount.New()
	ch := make(chan runner.LogLine)
	ctx, cancel := context.WithCancel(context.Background())
	out := c.Apply(ctx, ch)
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
