package sample

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/runner"
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

func TestNew_InvalidN_ReturnsError(t *testing.T) {
	_, err := New(0)
	if err == nil {
		t.Fatal("expected error for n=0")
	}
	_, err = New(-3)
	if err == nil {
		t.Fatal("expected error for n=-3")
	}
}

func TestNew_ValidN_NoError(t *testing.T) {
	_, err := New(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_N1_PassesAllLines(t *testing.T) {
	s, _ := New(1)
	lines := []runner.LogLine{
		{Service: "svc", Text: "a"},
		{Service: "svc", Text: "b"},
		{Service: "svc", Text: "c"},
	}
	out := collect(s.Apply(context.Background(), makeLineCh(lines)))
	if len(out) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(out))
	}
}

func TestApply_N2_PassesEverySecond(t *testing.T) {
	s, _ := New(2)
	lines := []runner.LogLine{
		{Service: "svc", Text: "1"},
		{Service: "svc", Text: "2"},
		{Service: "svc", Text: "3"},
		{Service: "svc", Text: "4"},
	}
	out := collect(s.Apply(context.Background(), makeLineCh(lines)))
	if len(out) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(out))
	}
	if out[0].Text != "2" || out[1].Text != "4" {
		t.Fatalf("unexpected lines: %+v", out)
	}
}

func TestApply_CountersIndependentPerService(t *testing.T) {
	s, _ := New(2)
	lines := []runner.LogLine{
		{Service: "a", Text: "a1"},
		{Service: "b", Text: "b1"},
		{Service: "a", Text: "a2"},
		{Service: "b", Text: "b2"},
	}
	out := collect(s.Apply(context.Background(), makeLineCh(lines)))
	// each service should emit its 2nd line
	if len(out) != 2 {
		t.Fatalf("expected 2 lines, got %d: %+v", len(out), out)
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	s, _ := New(1)
	src := make(chan runner.LogLine)
	ctx, cancel := context.WithCancel(context.Background())
	out := s.Apply(ctx, src)
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
