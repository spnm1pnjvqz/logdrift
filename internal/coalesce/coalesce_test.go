package coalesce_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/coalesce"
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

func TestNew_ZeroWindow_ReturnsError(t *testing.T) {
	_, err := coalesce.New(0)
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestNew_NegativeWindow_ReturnsError(t *testing.T) {
	_, err := coalesce.New(-time.Second)
	if err == nil {
		t.Fatal("expected error for negative window")
	}
}

func TestNew_ValidWindow_NoError(t *testing.T) {
	_, err := coalesce.New(50 * time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_CoalescesBurst(t *testing.T) {
	c, _ := coalesce.New(60 * time.Millisecond)
	in := makeLineCh([]runner.LogLine{
		{Service: "api", Text: "a"},
		{Service: "api", Text: "b"},
		{Service: "api", Text: "c"},
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out := c.Apply(ctx, in)
	var results []runner.LogLine
	for l := range out {
		results = append(results, l)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 merged line, got %d", len(results))
	}
	if results[0].Text != "a | b | c" {
		t.Errorf("unexpected merged text: %q", results[0].Text)
	}
}

func TestApply_DifferentServices_EmittedSeparately(t *testing.T) {
	c, _ := coalesce.New(60 * time.Millisecond)
	in := makeLineCh([]runner.LogLine{
		{Service: "api", Text: "x"},
		{Service: "db", Text: "y"},
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out := c.Apply(ctx, in)
	services := map[string]bool{}
	for l := range out {
		services[l.Service] = true
	}
	if !services["api"] || !services["db"] {
		t.Errorf("expected both services, got %v", services)
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	c, _ := coalesce.New(500 * time.Millisecond)
	ch := make(chan runner.LogLine)
	ctx, cancel := context.WithCancel(context.Background())
	out := c.Apply(ctx, ch)
	cancel()
	_, ok := <-out
	if ok {
		t.Error("expected channel to be closed after cancel")
	}
}
