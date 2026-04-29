package suppress

import (
	"context"
	"testing"

	"github.com/user/logdrift/internal/runner"
)

func mkLine(svc, text string) runner.LogLine {
	return runner.LogLine{Service: svc, Text: text}
}

func makeLineCh(lines []runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestNew_InvalidMaxReps_ReturnsError(t *testing.T) {
	_, err := New(0)
	if err == nil {
		t.Fatal("expected error for maxReps=0")
	}
}

func TestNew_ValidMaxReps_NoError(t *testing.T) {
	_, err := New(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAllow_FirstOccurrence_Passes(t *testing.T) {
	s, _ := New(2)
	if !s.Allow(mkLine("svc", "hello")) {
		t.Fatal("first occurrence should pass")
	}
}

func TestAllow_WithinLimit_Passes(t *testing.T) {
	s, _ := New(3)
	line := mkLine("svc", "repeat")
	for i := 0; i < 3; i++ {
		if !s.Allow(line) {
			t.Fatalf("occurrence %d should pass (limit=3)", i+1)
		}
	}
}

func TestAllow_ExceedsLimit_Dropped(t *testing.T) {
	s, _ := New(2)
	line := mkLine("svc", "repeat")
	s.Allow(line)
	s.Allow(line)
	if s.Allow(line) {
		t.Fatal("third occurrence should be dropped (limit=2)")
	}
}

func TestAllow_DifferentLine_ResetsCounter(t *testing.T) {
	s, _ := New(1)
	s.Allow(mkLine("svc", "a"))
	// second "a" should be dropped
	if s.Allow(mkLine("svc", "a")) {
		t.Fatal("second 'a' should be dropped")
	}
	// new line resets
	if !s.Allow(mkLine("svc", "b")) {
		t.Fatal("new line 'b' should pass")
	}
	// first "a" again — counter reset
	if !s.Allow(mkLine("svc", "a")) {
		t.Fatal("'a' after reset should pass")
	}
}

func TestAllow_IndependentPerService(t *testing.T) {
	s, _ := New(1)
	line1 := mkLine("svc1", "msg")
	line2 := mkLine("svc2", "msg")
	s.Allow(line1)
	s.Allow(line2)
	// second occurrence for each service should be dropped independently
	if s.Allow(line1) {
		t.Fatal("svc1 second occurrence should be dropped")
	}
	if s.Allow(line2) {
		t.Fatal("svc2 second occurrence should be dropped")
	}
}

func TestApply_FiltersExcessRepetitions(t *testing.T) {
	s, _ := New(2)
	lines := []runner.LogLine{
		mkLine("svc", "x"),
		mkLine("svc", "x"),
		mkLine("svc", "x"), // dropped
		mkLine("svc", "y"),
	}
	ctx := context.Background()
	out := Apply(ctx, s, makeLineCh(lines))
	var got []runner.LogLine
	for l := range out {
		got = append(got, l)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(got))
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	s, _ := New(1)
	ch := make(chan runner.LogLine) // never sends
	ctx, cancel := context.WithCancel(context.Background())
	out := Apply(ctx, s, ch)
	cancel()
	_, ok := <-out
	if ok {
		t.Fatal("channel should be closed after cancel")
	}
}
