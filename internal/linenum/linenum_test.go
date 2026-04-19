package linenum

import (
	"context"
	"strings"
	"testing"

	"github.com/user/logdrift/internal/runner"
)

func makeLine(service, text string) runner.LogLine {
	return runner.LogLine{Service: service, Text: text}
}

func makeLineCh(lines []runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestNew_DefaultPadding(t *testing.T) {
	s := New(0)
	if s.padding != 4 {
		t.Fatalf("expected default padding 4, got %d", s.padding)
	}
}

func TestNew_CustomPadding(t *testing.T) {
	s := New(6)
	if s.padding != 6 {
		t.Fatalf("expected padding 6, got %d", s.padding)
	}
}

func TestStamp_IncrementsPerService(t *testing.T) {
	s := New(1)
	l1 := s.Stamp(makeLine("svc", "hello"))
	l2 := s.Stamp(makeLine("svc", "world"))
	if !strings.HasPrefix(l1.Text, "1 ") {
		t.Fatalf("expected prefix '1 ', got %q", l1.Text)
	}
	if !strings.HasPrefix(l2.Text, "2 ") {
		t.Fatalf("expected prefix '2 ', got %q", l2.Text)
	}
}

func TestStamp_IndependentPerService(t *testing.T) {
	s := New(1)
	s.Stamp(makeLine("a", "x"))
	s.Stamp(makeLine("a", "x"))
	b := s.Stamp(makeLine("b", "x"))
	if !strings.HasPrefix(b.Text, "1 ") {
		t.Fatalf("service b counter should start at 1, got %q", b.Text)
	}
}

func TestApply_StampsAllLines(t *testing.T) {
	s := New(4)
	lines := []runner.LogLine{
		makeLine("web", "start"),
		makeLine("web", "stop"),
	}
	ch := makeLineCh(lines)
	out := s.Apply(context.Background(), ch)
	var got []runner.LogLine
	for l := range out {
		got = append(got, l)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
	if !strings.HasPrefix(got[0].Text, "0001 ") {
		t.Errorf("unexpected first line: %q", got[0].Text)
	}
	if !strings.HasPrefix(got[1].Text, "0002 ") {
		t.Errorf("unexpected second line: %q", got[1].Text)
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	s := New(1)
	ch := make(chan runner.LogLine) // never sends
	ctx, cancel := context.WithCancel(context.Background())
	out := s.Apply(ctx, ch)
	cancel()
	_, ok := <-out
	if ok {
		t.Fatal("expected channel to be closed after cancel")
	}
}
