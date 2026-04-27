package debounce

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/runner"
)

// TestIntegration_WindowResetOnRepeat verifies that repeated identical lines
// continuously reset the timer and only one line is emitted per burst.
func TestIntegration_WindowResetOnRepeat(t *testing.T) {
	window := 80 * time.Millisecond
	d, _ := New(window)

	ch := make(chan runner.LogLine, 10)
	out := Apply(context.Background(), d, ch)

	// Send three rapid duplicates.
	for i := 0; i < 3; i++ {
		ch <- line("svc", "burst")
	}
	// Wait for window to expire then send one more.
	time.Sleep(window + 20*time.Millisecond)
	ch <- line("svc", "burst")
	close(ch)

	var got []runner.LogLine
	for l := range out {
		got = append(got, l)
	}
	// Expect exactly 2: one at start of burst, one after window expired.
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
	for _, l := range got {
		if l.Text != "burst" {
			t.Errorf("unexpected text: %q", l.Text)
		}
	}
}
