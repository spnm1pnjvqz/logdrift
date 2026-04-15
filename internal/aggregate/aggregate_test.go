package aggregate

import (
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

func collectSummaries(ch <-chan Summary) []Summary {
	var out []Summary
	for s := range ch {
		out = append(out, s)
	}
	return out
}

func TestNew_InvalidWindow_ReturnsError(t *testing.T) {
	_, err := New(-1 * time.Second)
	if err == nil {
		t.Fatal("expected error for negative window")
	}
	if err != ErrInvalidWindow {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNew_ZeroWindow_ReturnsError(t *testing.T) {
	_, err := New(0)
	if err != ErrInvalidWindow {
		t.Fatalf("expected ErrInvalidWindow, got %v", err)
	}
}

func TestNew_ValidWindow_NoError(t *testing.T) {
	_, err := New(100 * time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_CountsPerService(t *testing.T) {
	lines := []runner.LogLine{
		{Service: "api", Text: "GET /"},
		{Service: "api", Text: "POST /login"},
		{Service: "worker", Text: "job started"},
	}
	agg, err := New(10 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	ch := makeLineCh(lines)
	summaries := collectSummaries(agg.Apply(ch))

	counts := map[string]int{}
	for _, s := range summaries {
		counts[s.Key] += s.Count
	}
	if counts["api"] != 2 {
		t.Errorf("api count: want 2, got %d", counts["api"])
	}
	if counts["worker"] != 1 {
		t.Errorf("worker count: want 1, got %d", counts["worker"])
	}
}

func TestApply_EmptyInput_NoSummaries(t *testing.T) {
	agg, _ := New(10 * time.Second)
	ch := makeLineCh(nil)
	summaries := collectSummaries(agg.Apply(ch))
	if len(summaries) != 0 {
		t.Errorf("expected no summaries, got %d", len(summaries))
	}
}

func TestApply_WindowEndSet(t *testing.T) {
	before := time.Now()
	lines := []runner.LogLine{{Service: "svc", Text: "msg"}}
	agg, _ := New(10 * time.Second)
	summaries := collectSummaries(agg.Apply(makeLineCh(lines)))
	after := time.Now()
	for _, s := range summaries {
		if s.WindowEnd.Before(before) || s.WindowEnd.After(after) {
			t.Errorf("WindowEnd %v outside [%v, %v]", s.WindowEnd, before, after)
		}
	}
}
