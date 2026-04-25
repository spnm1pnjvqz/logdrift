package ceiling

import (
	"testing"

	"github.com/user/logdrift/internal/runner"
)

func makeLineCh(lines []runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collect(ch <-chan runner.LogLine) []runner.LogLine {
	var out []runner.LogLine
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestNew_InvalidMax_ReturnsError(t *testing.T) {
	_, err := New(0)
	if err == nil {
		t.Fatal("expected error for max=0")
	}
}

func TestNew_ValidMax_NoError(t *testing.T) {
	_, err := New(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAllow_UnderCeiling_Passes(t *testing.T) {
	c, _ := New(3)
	l := runner.LogLine{Service: "svc", Text: "msg"}
	for i := 0; i < 3; i++ {
		if !c.Allow(l) {
			t.Fatalf("expected line %d to be allowed", i+1)
		}
	}
}

func TestAllow_AtCeiling_Drops(t *testing.T) {
	c, _ := New(2)
	l := runner.LogLine{Service: "svc", Text: "msg"}
	c.Allow(l)
	c.Allow(l)
	if c.Allow(l) {
		t.Fatal("expected third line to be dropped")
	}
}

func TestAllow_DifferentServices_Independent(t *testing.T) {
	c, _ := New(1)
	a := runner.LogLine{Service: "a", Text: "x"}
	b := runner.LogLine{Service: "b", Text: "x"}
	if !c.Allow(a) {
		t.Fatal("first line for 'a' should pass")
	}
	if !c.Allow(b) {
		t.Fatal("first line for 'b' should pass independently")
	}
	if c.Allow(a) {
		t.Fatal("second line for 'a' should be dropped")
	}
}

func TestApply_LimitsOutput(t *testing.T) {
	c, _ := New(2)
	lines := []runner.LogLine{
		{Service: "svc", Text: "1"},
		{Service: "svc", Text: "2"},
		{Service: "svc", Text: "3"},
	}
	got := collect(Apply(c, makeLineCh(lines)))
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
}

func TestApply_ClosesWhenInputClosed(t *testing.T) {
	c, _ := New(10)
	ch := make(chan runner.LogLine)
	close(ch)
	got := collect(Apply(c, ch))
	if len(got) != 0 {
		t.Fatalf("expected 0 lines, got %d", len(got))
	}
}
