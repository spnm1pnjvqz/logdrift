package retry_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/retry"
	"github.com/user/logdrift/internal/runner"
)

func makeClosingCh(lines []string, service string) func(ctx context.Context) (<-chan runner.LogLine, error) {
	calls := 0
	return func(ctx context.Context) (<-chan runner.LogLine, error) {
		ch := make(chan runner.LogLine)
		go func() {
			defer close(ch)
			if calls > 0 {
				return // second call returns empty channel immediately
			}
			calls++
			for _, l := range lines {
				ch <- runner.LogLine{Service: service, Text: l}
			}
		}()
		return ch, nil
	}
}

func TestNew_InvalidMaxAttempts(t *testing.T) {
	_, err := retry.New(retry.Config{MaxAttempts: 0})
	if err == nil {
		t.Fatal("expected error for MaxAttempts=0")
	}
}

func TestNew_ValidConfig(t *testing.T) {
	_, err := retry.New(retry.Config{MaxAttempts: 3, Delay: time.Millisecond})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_ReceivesLines(t *testing.T) {
	r, _ := retry.New(retry.Config{MaxAttempts: 1, Delay: 0})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	factory := makeClosingCh([]string{"hello", "world"}, "svc")
	ch, err := r.Apply(ctx, factory)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got []string
	for line := range ch {
		got = append(got, line.Text)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
}

func TestApply_RetriesOnClose(t *testing.T) {
	r, _ := retry.New(retry.Config{MaxAttempts: 2, Delay: time.Millisecond})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	factory := makeClosingCh([]string{"a"}, "svc")
	ch, err := r.Apply(ctx, factory)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got []string
	for line := range ch {
		got = append(got, line.Text)
	}
	// first call emits "a", second call emits nothing, then stops
	if len(got) != 1 || got[0] != "a" {
		t.Fatalf("unexpected lines: %v", got)
	}
}

func TestApply_CancelStopsOutput(t *testing.T) {
	r, _ := retry.New(retry.Config{MaxAttempts: 5, Delay: time.Second})
	ctx, cancel := context.WithCancel(context.Background())

	blocking := func(ctx context.Context) (<-chan runner.LogLine, error) {
		ch := make(chan runner.LogLine)
		return ch, nil
	}
	ch, _ := r.Apply(ctx, blocking)
	cancel()
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("channel did not close after cancel")
	}
}
