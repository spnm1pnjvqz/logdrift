package pause_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/user/logdrift/internal/pause"
	"github.com/user/logdrift/internal/runner"
)

// TestApply_OrderPreservedAcrossPauseResume verifies that lines emitted
// before, during, and after a pause arrive in the correct order.
func TestApply_OrderPreservedAcrossPauseResume(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ctrl := pause.New()
	in := make(chan runner.LogLine, 10)

	for i := 0; i < 3; i++ {
		in <- runner.LogLine{Service: "s", Text: string(rune('a' + i))}
	}

	out := pause.Apply(ctx, ctrl, in)

	// collect first 3 lines before pausing
	var got []string
	for i := 0; i < 3; i++ {
		select {
		case l := <-out:
			got = append(got, l.Text)
		case <-ctx.Done():
			t.Fatal("timeout collecting pre-pause lines")
		}
	}

	ctrl.Pause()

	// send lines while paused
	for i := 3; i < 6; i++ {
		in <- runner.LogLine{Service: "s", Text: string(rune('a' + i))}
	}
	close(in)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(60 * time.Millisecond)
		ctrl.Resume()
	}()

	for l := range out {
		got = append(got, l.Text)
	}
	wg.Wait()

	expected := []string{"a", "b", "c", "d", "e", "f"}
	if len(got) != len(expected) {
		t.Fatalf("expected %d lines, got %d: %v", len(expected), len(got), got)
	}
	for i, v := range expected {
		if got[i] != v {
			t.Errorf("position %d: want %q, got %q", i, v, got[i])
		}
	}
}
