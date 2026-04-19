package indent

import (
	"context"
	"testing"
	"time"

	"github.com/celrenshaw/logdrift/internal/runner"
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
	_, err := New("  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStamp_PrependsPrefix(t *testing.T) {
	in, _ := New(">> ")
	l := runner.LogLine{Service: "svc", Text: "hello"}
	got := in.Stamp(l)
	if got.Text != ">> hello" {
		t.Errorf("got %q, want %q", got.Text, ">> hello")
	}
	if got.Service != "svc" {
		t.Errorf("service changed: got %q", got.Service)
	}
}

func TestApply_PrependsToAllLines(t *testing.T) {
	in, _ := New("\t")
	lines := []runner.LogLine{
		{Service: "a", Text: "line1"},
		{Service: "a", Text: "line2"},
	}
	src := makeLineCh(lines)
	out := in.Apply(context.Background(), src)
	var got []runner.LogLine
	for l := range out {
		got = append(got, l)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
	for _, l := range got {
		if len(l.Text) == 0 || l.Text[0] != '\t' {
			t.Errorf("line not indented: %q", l.Text)
		}
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	in, _ := New("  ")
	ch := make(chan runner.LogLine)
	ctx, cancel := context.WithCancel(context.Background())
	out := in.Apply(ctx, ch)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for channel close")
	}
}
