package differ

import (
	"context"
	"testing"
	"time"
)

func makeLineCh(lines []Line) <-chan Line {
	ch := make(chan Line, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestPipeline_EmitsAllEvents(t *testing.T) {
	d := New(DiffModeUniq, 0)
	p := NewPipeline(d)

	lines := []Line{
		{Service: "a", Text: "line one", Timestamp: time.Now()},
		{Service: "b", Text: "line two", Timestamp: time.Now()},
	}
	in := makeLineCh(lines)

	ctx := context.Background()
	out := p.Run(ctx, in)

	var events []Event
	for e := range out {
		events = append(events, e)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
}

func TestPipeline_FirstLineDrift_SecondNot(t *testing.T) {
	d := New(DiffModeUniq, 0)
	p := NewPipeline(d)

	lines := []Line{
		{Service: "a", Text: "shared message", Timestamp: time.Now()},
		{Service: "b", Text: "shared message", Timestamp: time.Now()},
	}
	in := makeLineCh(lines)

	out := p.Run(context.Background(), in)

	events := collectEvents(out)
	if !events[0].Drift {
		t.Error("first occurrence should be drift")
	}
	if events[1].Drift {
		t.Error("second occurrence of same line should not be drift")
	}
}

func TestPipeline_CancelStopsOutput(t *testing.T) {
	d := New(DiffModeUniq, 0)
	p := NewPipeline(d)

	in := make(chan Line) // never sends
	ctx, cancel := context.WithCancel(context.Background())
	out := p.Run(ctx, in)

	cancel()

	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for channel close")
	}
}

func collectEvents(ch <-chan Event) []Event {
	var out []Event
	for e := range ch {
		out = append(out, e)
	}
	return out
}
