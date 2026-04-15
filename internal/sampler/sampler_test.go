package sampler_test

import (
	"testing"
	"time"

	"github.com/user/logdrift/internal/runner"
	"github.com/user/logdrift/internal/sampler"
)

func makeLineCh(lines []string) <-chan runner.Line {
	ch := make(chan runner.Line, len(lines))
	for _, l := range lines {
		ch <- runner.Line{Service: "svc", Text: l}
	}
	close(ch)
	return ch
}

func collectLines(ch <-chan runner.Line, timeout time.Duration) []runner.Line {
	var out []runner.Line
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case l, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, l)
		case <-timer.C:
			return out
		}
	}
}

func TestValidate_InvalidN(t *testing.T) {
	cfg := sampler.Config{N: 0}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for N=0, got nil")
	}
}

func TestValidate_ValidN(t *testing.T) {
	cfg := sampler.Config{N: 1}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_N1_AllLines(t *testing.T) {
	input := []string{"a", "b", "c", "d"}
	out := sampler.Apply(sampler.Config{N: 1}, makeLineCh(input))
	got := collectLines(out, 2*time.Second)
	if len(got) != len(input) {
		t.Fatalf("expected %d lines, got %d", len(input), len(got))
	}
}

func TestApply_N2_HalfLines(t *testing.T) {
	input := []string{"a", "b", "c", "d"}
	out := sampler.Apply(sampler.Config{N: 2}, makeLineCh(input))
	got := collectLines(out, 2*time.Second)
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
	if got[0].Text != "b" || got[1].Text != "d" {
		t.Fatalf("unexpected lines: %+v", got)
	}
}

func TestApply_ClosesOutput(t *testing.T) {
	in := makeLineCh(nil)
	out := sampler.Apply(sampler.Config{N: 1}, in)
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for channel close")
	}
}
