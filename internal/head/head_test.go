package head

import (
	"context"
	"testing"

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

func TestNew_InvalidN_ReturnsError(t *testing.T) {
	_, err := New(0)
	if err == nil {
		t.Fatal("expected error for n=0")
	}
}

func TestNew_ValidN_NoError(t *testing.T) {
	_, err := New(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_LimitsPerService(t *testing.T) {
	lim, _ := New(2)
	input := []runner.LogLine{
		{Service: "svc", Text: "a"},
		{Service: "svc", Text: "b"},
		{Service: "svc", Text: "c"},
	}
	got := collect(lim.Apply(context.Background(), makeLineCh(input)))
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
}

func TestApply_IndependentPerService(t *testing.T) {
	lim, _ := New(1)
	input := []runner.LogLine{
		{Service: "a", Text: "1"},
		{Service: "b", Text: "2"},
		{Service: "a", Text: "3"},
	}
	got := collect(lim.Apply(context.Background(), makeLineCh(input)))
	if len(got) != 2 {
		t.Fatalf("expected 2 lines (one per service), got %d", len(got))
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	lim, _ := New(100)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	in := make(chan runner.LogLine)
	close(in)
	got := collect(lim.Apply(ctx, in))
	if len(got) != 0 {
		t.Fatalf("expected 0 lines after cancel, got %d", len(got))
	}
}
