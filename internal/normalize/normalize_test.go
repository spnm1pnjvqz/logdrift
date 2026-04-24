package normalize_test

import (
	"context"
	"testing"
	"time"

	"github.com/myorg/logdrift/internal/normalize"
	"github.com/myorg/logdrift/internal/runner"
)

func line(svc, text string) runner.LogLine {
	return runner.LogLine{Service: svc, Text: text}
}

func makeLineCh(lines ...runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestNew_NoOptions_ReturnsError(t *testing.T) {
	_, err := normalize.New(normalize.Options{})
	if err == nil {
		t.Fatal("expected error for empty options")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := normalize.New(normalize.Options{Lowercase: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_Lowercase(t *testing.T) {
	n, _ := normalize.New(normalize.Options{Lowercase: true})
	got := n.Apply(line("svc", "Hello WORLD"))
	if got.Text != "hello world" {
		t.Fatalf("want %q got %q", "hello world", got.Text)
	}
}

func TestApply_CollapseSpaces(t *testing.T) {
	n, _ := normalize.New(normalize.Options{CollapseSpaces: true})
	got := n.Apply(line("svc", "foo  bar\t\tbaz"))
	if got.Text != "foo bar baz" {
		t.Fatalf("want %q got %q", "foo bar baz", got.Text)
	}
}

func TestApply_Trim(t *testing.T) {
	n, _ := normalize.New(normalize.Options{Trim: true})
	got := n.Apply(line("svc", "  hello  "))
	if got.Text != "hello" {
		t.Fatalf("want %q got %q", "hello", got.Text)
	}
}

func TestApply_AllOptions_Combined(t *testing.T) {
	n, _ := normalize.New(normalize.Options{Lowercase: true, CollapseSpaces: true, Trim: true})
	got := n.Apply(line("svc", "  FOO   BAR  "))
	if got.Text != "foo bar" {
		t.Fatalf("want %q got %q", "foo bar", got.Text)
	}
}

func TestApply_PreservesService(t *testing.T) {
	n, _ := normalize.New(normalize.Options{Lowercase: true})
	got := n.Apply(line("MyService", "TEXT"))
	if got.Service != "MyService" {
		t.Fatalf("service should be unchanged, got %q", got.Service)
	}
}

func TestApplyAll_ClosesWhenInputClosed(t *testing.T) {
	n, _ := normalize.New(normalize.Options{Trim: true})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	in := makeLineCh(line("a", " hello "), line("b", " world "))
	out := n.ApplyAll(ctx, in)

	var results []runner.LogLine
	for l := range out {
		results = append(results, l)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(results))
	}
	if results[0].Text != "hello" || results[1].Text != "world" {
		t.Fatalf("unexpected texts: %v", results)
	}
}

func TestApplyAll_Cancel_StopsOutput(t *testing.T) {
	n, _ := normalize.New(normalize.Options{Trim: true})
	ctx, cancel := context.WithCancel(context.Background())

	blocking := make(chan runner.LogLine)
	out := n.ApplyAll(ctx, blocking)

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
