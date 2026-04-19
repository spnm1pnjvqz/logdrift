package sequence

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/user/logdrift/internal/runner"
)

func makeLineCh(texts ...string) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(texts))
	for _, t := range texts {
		ch <- runner.LogLine{Service: "svc", Text: t}
	}
	close(ch)
	return ch
}

func TestNew_EmptyPrefix_UsesDefault(t *testing.T) {
	s, err := New("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.prefix != "#" {
		t.Fatalf("expected default prefix '#', got %q", s.prefix)
	}
}

func TestNew_CustomPrefix(t *testing.T) {
	s, err := New("seq:")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.prefix != "seq:" {
		t.Fatalf("expected 'seq:', got %q", s.prefix)
	}
}

func TestStamp_IncrementsCounter(t *testing.T) {
	s, _ := New("#")
	l1 := s.Stamp(runner.LogLine{Text: "hello"})
	l2 := s.Stamp(runner.LogLine{Text: "world"})
	if !strings.Contains(l1.Text, "[#1]") {
		t.Errorf("expected [#1] in %q", l1.Text)
	}
	if !strings.Contains(l2.Text, "[#2]") {
		t.Errorf("expected [#2] in %q", l2.Text)
	}
}

func TestStamp_PreservesService(t *testing.T) {
	s, _ := New("#")
	original := runner.LogLine{Service: "mysvc", Text: "hello"}
	stamped := s.Stamp(original)
	if stamped.Service != original.Service {
		t.Errorf("expected service %q, got %q", original.Service, stamped.Service)
	}
}

func TestApply_StampsAllLines(t *testing.T) {
	s, _ := New("#")
	ctx := context.Background()
	in := makeLineCh("a", "b", "c")
	out := s.Apply(ctx, in)
	var lines []runner.LogLine
	for l := range out {
		lines = append(lines, l)
	}
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	for i, l := range lines {
		tag := fmt.Sprintf("[#%d]", i+1)
		if !strings.Contains(l.Text, tag) {
			t.Errorf("line %d: expected tag %s in %q", i+1, tag, l.Text)
		}
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	s, _ := New("#")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	inf := make(chan runner.LogLine) // never closes
	out := s.Apply(ctx, inf)
	_, ok := <-out
	if ok {
		t.Error("expected channel to be closed after cancel")
	}
}
