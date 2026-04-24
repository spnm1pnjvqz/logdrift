package bracket

import (
	"context"
	"testing"

	"github.com/user/logdrift/internal/runner"
)

func collectLines(ch <-chan runner.LogLine) []runner.LogLine {
	var out []runner.LogLine
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestApply_ClosesWhenInputClosed(t *testing.T) {
	b, _ := New("{", "}")
	src := makeLineCh() // empty, already closed
	out := b.Apply(context.Background(), src)
	lines := collectLines(out)
	if len(lines) != 0 {
		t.Errorf("expected 0 lines, got %d", len(lines))
	}
}

func TestApply_OnlyOpen_NoClose(t *testing.T) {
	b, _ := New(">> ", "")
	src := makeLineCh(makeLine("svc", "msg"))
	lines := collectLines(b.Apply(context.Background(), src))
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0].Text != ">> msg" {
		t.Errorf("got %q", lines[0].Text)
	}
}

func TestApply_OnlyClose_NoOpen(t *testing.T) {
	b, _ := New("", " <<")
	src := makeLineCh(makeLine("svc", "msg"))
	lines := collectLines(b.Apply(context.Background(), src))
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0].Text != "msg <<" {
		t.Errorf("got %q", lines[0].Text)
	}
}

func TestApply_MultipleServices_EachWrapped(t *testing.T) {
	b, _ := New("[", "]")
	src := makeLineCh(
		makeLine("api", "start"),
		makeLine("db", "query"),
		makeLine("api", "end"),
	)
	lines := collectLines(b.Apply(context.Background(), src))
	expected := []string{"[start]", "[query]", "[end]"}
	for i, l := range lines {
		if l.Text != expected[i] {
			t.Errorf("line %d: got %q, want %q", i, l.Text, expected[i])
		}
	}
}
