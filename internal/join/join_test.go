package join

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

func collect(ctx context.Context, ch <-chan runner.LogLine) []runner.LogLine {
	var out []runner.LogLine
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestNew_EmptySeparator_ReturnsError(t *testing.T) {
	_, err := New("", 2)
	if err == nil {
		t.Fatal("expected error for empty separator")
	}
}

func TestNew_MaxLinesBelowTwo_ReturnsError(t *testing.T) {
	_, err := New(" ", 1)
	if err == nil {
		t.Fatal("expected error for maxLines < 2")
	}
}

func TestNew_ValidArgs_NoError(t *testing.T) {
	_, err := New(" | ", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_JoinsUpToMaxLines(t *testing.T) {
	j, _ := New(" | ", 3)
	lines := []runner.LogLine{
		{Service: "svc", Text: "a"},
		{Service: "svc", Text: "b"},
		{Service: "svc", Text: "c"},
	}
	ctx := context.Background()
	results := collect(ctx, j.Apply(ctx, makeLineCh(lines)))
	if len(results) != 1 {
		t.Fatalf("expected 1 joined line, got %d", len(results))
	}
	if results[0].Text != "a | b | c" {
		t.Errorf("unexpected joined text: %q", results[0].Text)
	}
}

func TestApply_FlushesOnServiceChange(t *testing.T) {
	j, _ := New("+", 10)
	lines := []runner.LogLine{
		{Service: "a", Text: "x"},
		{Service: "a", Text: "y"},
		{Service: "b", Text: "z"},
	}
	ctx := context.Background()
	results := collect(ctx, j.Apply(ctx, makeLineCh(lines)))
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Text != "x+y" {
		t.Errorf("unexpected text: %q", results[0].Text)
	}
	if results[1].Text != "z" {
		t.Errorf("unexpected text: %q", results[1].Text)
	}
}

func TestApply_FlushesWhenMaxLinesReached(t *testing.T) {
	// When the number of buffered lines reaches maxLines, the joiner should
	// emit a joined line and start a new group rather than accumulating further.
	j, _ := New("-", 2)
	lines := []runner.LogLine{
		{Service: "svc", Text: "1"},
		{Service: "svc", Text: "2"},
		{Service: "svc", Text: "3"},
	}
	ctx := context.Background()
	results := collect(ctx, j.Apply(ctx, makeLineCh(lines)))
	if len(results) != 2 {
		t.Fatalf("expected 2 results after maxLines flush, got %d", len(results))
	}
	if results[0].Text != "1-2" {
		t.Errorf("unexpected first group text: %q", results[0].Text)
	}
	if results[1].Text != "3" {
		t.Errorf("unexpected second group text: %q", results[1].Text)
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	j, _ := New("-", 2)
	ch := make(chan runner.LogLine)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	out := j.Apply(ctx, ch)
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to close after cancel")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for channel close")
	}
}
