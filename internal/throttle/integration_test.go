package throttle_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/runner"
	"github.com/user/logdrift/internal/throttle"
)

// TestApply_RateIsRespected verifies that the throttle introduces measurable
// delay between lines when linesPerSec is small.
func TestApply_RateIsRespected(t *testing.T) {
	const lps = 20
	th, err := throttle.New(lps)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	const n = 5
	ch := make(chan runner.LogLine, n)
	for i := 0; i < n; i++ {
		ch <- runner.LogLine{Service: "svc", Text: "x"}
	}
	close(ch)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	out := th.Apply(ctx, ch)
	count := 0
	for range out {
		count++
	}
	elapsed := time.Since(start)

	if count != n {
		t.Fatalf("expected %d lines, got %d", n, count)
	}
	// n lines at lps lines/sec should take at least (n-1)/lps seconds
	minExpected := time.Duration(n-1) * time.Second / time.Duration(lps)
	if elapsed < minExpected {
		t.Errorf("elapsed %v < expected minimum %v — throttle not applied", elapsed, minExpected)
	}
}
