package jsonformat

import (
	"context"
	"testing"

	"github.com/user/logdrift/internal/runner"
)

func line(text string) runner.LogLine {
	return runner.LogLine{Service: "svc", Text: text}
}

func TestNew_DefaultIndent(t *testing.T) {
	f := New("")
	if f.indent != "  " {
		t.Fatalf("expected two spaces, got %q", f.indent)
	}
}

func TestFormat_NonJSON_Unchanged(t *testing.T) {
	f := New("")
	l := line("plain log line")
	got := f.Format(l)
	if got.Text != "plain log line" {
		t.Fatalf("expected unchanged, got %q", got.Text)
	}
}

func TestFormat_ValidJSON_PrettyPrinted(t *testing.T) {
	f := New("  ")
	l := line(`{"level":"info","msg":"hello"}`)
	got := f.Format(l)
	if got.Text == l.Text {
		t.Fatal("expected pretty-printed output to differ from compact input")
	}
	if len(got.Text) < len(l.Text) {
		t.Fatal("pretty-printed output should be longer")
	}
}

func TestFormat_InvalidJSON_Unchanged(t *testing.T) {
	f := New("")
	input := `{bad json`
	l := line(input)
	got := f.Format(l)
	if got.Text != input {
		t.Fatalf("expected unchanged, got %q", got.Text)
	}
}

func TestFormat_JSONArray_PrettyPrinted(t *testing.T) {
	f := New("\t")
	l := line(`[1,2,3]`)
	got := f.Format(l)
	if got.Text == l.Text {
		t.Fatal("expected pretty-printed array")
	}
}

func TestFormat_PreservesService(t *testing.T) {
	f := New("")
	l := runner.LogLine{Service: "api", Text: `{"x":1}`}
	got := f.Format(l)
	if got.Service != "api" {
		t.Fatalf("service not preserved, got %q", got.Service)
	}
}

func makeLineCh(lines ...runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestApply_FormatsAndCloses(t *testing.T) {
	f := New("")
	in := makeLineCh(
		line(`{"a":1}`),
		line("not json"),
	)
	out := f.Apply(context.Background(), in)
	var results []runner.LogLine
	for l := range out {
		results = append(results, l)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(results))
	}
	if results[1].Text != "not json" {
		t.Fatalf("non-JSON line should be unchanged, got %q", results[1].Text)
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	f := New("")
	ch := make(chan runner.LogLine)
	ctx, cancel := context.WithCancel(context.Background())
	out := f.Apply(ctx, ch)
	cancel()
	for range out {
	}
}
