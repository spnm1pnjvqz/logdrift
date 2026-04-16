package throttle_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/runner"
	"github.com/user/logdrift/internal/throttle"
)

func makeLineCh(lines []string) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- runner.LogLine{Service: "svc", Text: l}
	}
	close(ch)
	return ch
}

func TestNew_InvalidRate_ReturnsError(t *testing.T) {
	_, err := throttle.New(0)
	if err == nil {
		t.Fatal("expected error for linesPerSec=0")
	}
}

func TestNew_ValidRate_NoError(t *testing.T) {
	_, err := throttle.New(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_PassesAllLines(t *testing.T) {
	th, _ := throttle.New(1000) // fast enough for tests
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	input := []string{"alpha", "beta", "gamma"}
	ch := makeLineCh(input)
	out := th.Apply(ctx, ch)

	var got []string
	for line := range out {
		got = append(got, line.Text)
	}
	if len(got) != len(input) {
		t.Fatalf("expected %d lines, got %d", len(input), len(got))
	}
	for i, v := range input {
		if got[i] != v {
			t.Errorf("line %d: want %q got %q", i, v, got[i])
		}
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	th, _ := throttle.New(1) // 1 line/sec — slow
	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan runner.LogLine, 100)
	for i := 0; i < 100; i++ {
		ch <- runner.LogLine{Service: "svc", Text: "line"}
	}

	out := th.Apply(ctx, ch)
	// receive at most one then cancel
	<-out
	cancel()

	timeout := time.After(2 * time.Second)
	for {
		select {
		case _, ok := <-out:
			if !ok {
				return
			}
		case <-timeout:
			t.Fatal("channel not closed after cancel")
		}
	}
}
