package prefix

import (
	"context"
	"testing"

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

func TestNew_EmptyPrefix_ReturnsError(t *testing.T) {
	_, err := New("")
	if err == nil {
		t.Fatal("expected error for empty prefix")
	}
}

func TestNew_ValidPrefix_NoError(t *testing.T) {
	_, err := New("[svc] ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_PrependsToAllLines(t *testing.T) {
	p, _ := New("[app] ")
	lines := []runner.LogLine{
		{Service: "svc", Text: "hello"},
		{Service: "svc", Text: "world"},
	}
	ctx := context.Background()
	out := p.Apply(ctx, makeLineCh(lines))

	var got []runner.LogLine
	for l := range out {
		got = append(got, l)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
	if got[0].Text != "[app] hello" {
		t.Errorf("expected '[app] hello', got %q", got[0].Text)
	}
	if got[1].Text != "[app] world" {
		t.Errorf("expected '[app] world', got %q", got[1].Text)
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	p, _ := New(">> ")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ch := make(chan runner.LogLine) // never sends
	out := p.Apply(ctx, ch)
	_, open := <-out
	if open {
		t.Error("expected channel to be closed after cancel")
	}
}
