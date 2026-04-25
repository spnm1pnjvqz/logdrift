package offset_test

import (
	"context"
	"testing"

	"github.com/user/logdrift/internal/offset"
	"github.com/user/logdrift/internal/runner"
)

func makeLine(service, text string) runner.LogLine {
	return runner.LogLine{Service: service, Text: text}
}

func makeLineCh(lines ...runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestNew_DefaultFormat_NoError(t *testing.T) {
	_, err := offset.New("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNew_CustomFormat_NoError(t *testing.T) {
	_, err := offset.New("%d|%s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStamp_OffsetStartsAtZero(t *testing.T) {
	s, _ := offset.New("[%d] %s")
	got := s.Stamp(makeLine("svc", "hello"))
	want := "[0] hello"
	if got.Text != want {
		t.Errorf("got %q, want %q", got.Text, want)
	}
}

func TestStamp_OffsetIncrementsPerService(t *testing.T) {
	s, _ := offset.New("%d|%s")
	s.Stamp(makeLine("svc", "hello")) // len=5 +1 = 6
	s.Stamp(makeLine("svc", "world")) // offset should be 6
	third := s.Stamp(makeLine("svc", "!"))  // offset should be 12
	if third.Text != "12|!" {
		t.Errorf("got %q, want %q", third.Text, "12|!")
	}
}

func TestStamp_IndependentPerService(t *testing.T) {
	s, _ := offset.New("%d|%s")
	s.Stamp(makeLine("a", "hello"))
	got := s.Stamp(makeLine("b", "first"))
	if got.Text != "0|first" {
		t.Errorf("service b offset should start at 0, got %q", got.Text)
	}
}

func TestCurrent_ReflectsAccumulatedBytes(t *testing.T) {
	s, _ := offset.New("%d|%s")
	s.Stamp(makeLine("svc", "abc")) // 3+1=4
	if got := s.Current("svc"); got != 4 {
		t.Errorf("expected 4, got %d", got)
	}
}

func TestReset_ZeroesOffset(t *testing.T) {
	s, _ := offset.New("%d|%s")
	s.Stamp(makeLine("svc", "hello"))
	s.Reset("svc")
	if got := s.Current("svc"); got != 0 {
		t.Errorf("expected 0 after reset, got %d", got)
	}
}

func TestApply_StampsAllLines(t *testing.T) {
	s, _ := offset.New("%d|%s")
	src := makeLineCh(
		makeLine("svc", "aa"),
		makeLine("svc", "bb"),
	)
	ctx := context.Background()
	out := s.Apply(ctx, src)

	first := <-out
	if first.Text != "0|aa" {
		t.Errorf("first: got %q, want %q", first.Text, "0|aa")
	}
	second := <-out
	if second.Text != "3|bb" {
		t.Errorf("second: got %q, want %q", second.Text, "3|bb")
	}
	_, ok := <-out
	if ok {
		t.Error("channel should be closed")
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	s, _ := offset.New("")
	ch := make(chan runner.LogLine) // never sends
	ctx, cancel := context.WithCancel(context.Background())
	out := s.Apply(ctx, ch)
	cancel()
	_, ok := <-out
	if ok {
		t.Error("expected channel to be closed after cancel")
	}
}
