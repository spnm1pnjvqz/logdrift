package strip

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/runner"
)

func TestNew_NoOptions_ReturnsError(t *testing.T) {
	_, err := New(Options{})
	if err == nil {
		t.Fatal("expected error for empty options")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := New(Options{ANSI: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_ANSI_RemovesEscapeCodes(t *testing.T) {
	s, _ := New(Options{ANSI: true})
	got := s.Apply("\x1b[31mhello\x1b[0m")
	if got != "hello" {
		t.Fatalf("expected 'hello', got %q", got)
	}
}

func TestApply_Whitespace_Trimmed(t *testing.T) {
	s, _ := New(Options{Whitespace: true})
	got := s.Apply("  hello  ")
	if got != "hello" {
		t.Fatalf("expected 'hello', got %q", got)
	}
}

func TestApply_Both_ANSIAndWhitespace(t *testing.T) {
	s, _ := New(Options{ANSI: true, Whitespace: true})
	got := s.Apply("  \x1b[32mworld\x1b[0m  ")
	if got != "world" {
		t.Fatalf("expected 'world', got %q", got)
	}
}

func makeLineCh(lines []runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestStream_StripsAllLines(t *testing.T) {
	s, _ := New(Options{ANSI: true, Whitespace: true})
	input := []runner.LogLine{
		{Service: "svc", Text: "  \x1b[31mfoo\x1b[0m  "},
		{Service: "svc", Text: "\x1b[32mbar\x1b[0m"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out := s.Stream(ctx, makeLineCh(input))
	var got []string
	for l := range out {
		got = append(got, l.Text)
	}
	if len(got) != 2 || got[0] != "foo" || got[1] != "bar" {
		t.Fatalf("unexpected output: %v", got)
	}
}

func TestStream_Cancel_StopsOutput(t *testing.T) {
	s, _ := New(Options{ANSI: true})
	ch := make(chan runner.LogLine)
	ctx, cancel := context.WithCancel(context.Background())
	out := s.Stream(ctx, ch)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for channel close")
	}
}
