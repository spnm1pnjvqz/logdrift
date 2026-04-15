package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/ratelimit"
	"github.com/user/logdrift/internal/runner"
)

func makeLineCh(lines []runner.Line) <-chan runner.Line {
	ch := make(chan runner.Line, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestNew_NegativeRate_ReturnsError(t *testing.T) {
	_, err := ratelimit.New(-1)
	if err == nil {
		t.Fatal("expected error for negative linesPerSec, got nil")
	}
}

func TestNew_ZeroRate_NoError(t *testing.T) {
	_, err := ratelimit.New(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_Unlimited_PassesAllLines(t *testing.T) {
	l, _ := ratelimit.New(0)
	input := []runner.Line{
		{Service: "svc", Text: "line1"},
		{Service: "svc", Text: "line2"},
		{Service: "svc", Text: "line3"},
	}
	ctx := context.Background()
	out := l.Apply(ctx, makeLineCh(input))

	var got []runner.Line
	for line := range out {
		got = append(got, line)
	}
	if len(got) != len(input) {
		t.Fatalf("expected %d lines, got %d", len(input), len(got))
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	l, _ := ratelimit.New(1) // 1 line/sec so it blocks between lines

	// Provide a large buffered channel that never closes on its own.
	ch := make(chan runner.Line, 100)
	for i := 0; i < 50; i++ {
		ch <- runner.Line{Service: "svc", Text: "x"}
	}

	ctx, cancel := context.WithCancel(context.Background())
	out := l.Apply(ctx, ch)

	// Cancel almost immediately.
	time.AfterFunc(50*time.Millisecond, cancel)

	var count int
	for range out {
		count++
	}
	// At 1 line/sec we expect far fewer than 50 lines in 50ms.
	if count >= 10 {
		t.Fatalf("rate limiter did not throttle: got %d lines", count)
	}
}

func TestApply_RateLimited_ThrottlesOutput(t *testing.T) {
	l, _ := ratelimit.New(10) // 10 lines/sec → 100ms per line

	input := []runner.Line{
		{Service: "a", Text: "1"},
		{Service: "a", Text: "2"},
		{Service: "a", Text: "3"},
	}
	ctx := context.Background()
	start := time.Now()
	out := l.Apply(ctx, makeLineCh(input))
	for range out {
	}
	elapsed := time.Since(start)

	// 3 lines at 10/sec should take at least ~200ms (2 inter-line gaps).
	if elapsed < 150*time.Millisecond {
		t.Fatalf("expected throttling, but finished in %v", elapsed)
	}
}
