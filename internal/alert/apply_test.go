package alert

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/runner"
)

func makeLogLineCh(lines []runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestApply_EmitsMatchingEvents(t *testing.T) {
	a, _ := New(map[string]string{"err": "ERROR"})
	lines := []runner.LogLine{
		{Service: "api", Text: "INFO started"},
		{Service: "api", Text: "ERROR disk full"},
		{Service: "api", Text: "INFO done"},
	}
	ctx := context.Background()
	out := Apply(ctx, a, makeLogLineCh(lines))
	var evs []Event
	for ev := range out {
		evs = append(evs, ev)
	}
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Service != "api" || evs[0].Rule != "err" {
		t.Fatalf("unexpected event: %+v", evs[0])
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	a, _ := New(map[string]string{"err": "ERROR"})
	ch := make(chan runner.LogLine) // never sends
	ctx, cancel := context.WithCancel(context.Background())
	out := Apply(ctx, a, ch)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for channel close")
	}
}
