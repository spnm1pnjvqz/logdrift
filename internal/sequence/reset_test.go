package sequence

import (
	"testing"

	"github.com/user/logdrift/internal/runner"
)

func TestCurrent_ZeroBeforeAnyStamp(t *testing.T) {
	s, _ := New("#")
	if s.Current() != 0 {
		t.Fatalf("expected 0, got %d", s.Current())
	}
}

func TestCurrent_ReflectsStampCount(t *testing.T) {
	s, _ := New("#")
	s.Stamp(runner.LogLine{Text: "x"})
	s.Stamp(runner.LogLine{Text: "y"})
	if s.Current() != 2 {
		t.Fatalf("expected 2, got %d", s.Current())
	}
}

func TestReset_ZeroesCounter(t *testing.T) {
	s, _ := New("#")
	s.Stamp(runner.LogLine{Text: "x"})
	s.Reset()
	if s.Current() != 0 {
		t.Fatalf("expected 0 after reset, got %d", s.Current())
	}
}

func TestReset_StampAfterResetStartsAt1(t *testing.T) {
	s, _ := New("#")
	s.Stamp(runner.LogLine{Text: "x"})
	s.Reset()
	l := s.Stamp(runner.LogLine{Text: "y"})
	if s.Current() != 1 {
		t.Fatalf("expected counter 1 after reset+stamp, got %d", s.Current())
	}
	if l.Text != "y [#1]" {
		t.Fatalf("unexpected text: %q", l.Text)
	}
}
