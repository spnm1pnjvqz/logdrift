package retry_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/logdrift/internal/retry"
	"github.com/user/logdrift/internal/runner"
)

func TestApply_ExhaustsAllAttempts(t *testing.T) {
	const maxAttempts = 4
	var calls atomic.Int32

	factory := func(ctx context.Context) (<-chan runner.LogLine, error) {
		calls.Add(1)
		ch := make(chan runner.LogLine)
		close(ch) // always closes immediately
		return ch, nil
	}

	r, err := retry.New(retry.Config{MaxAttempts: maxAttempts, Delay: time.Millisecond})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch, err := r.Apply(ctx, factory)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// drain
	for range ch {
	}

	if got := int(calls.Load()); got != maxAttempts {
		t.Fatalf("expected %d factory calls, got %d", maxAttempts, got)
	}
}
