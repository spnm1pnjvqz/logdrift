package runner

import (
	"context"
	"sort"
	"testing"
)

func makeStaticCh(lines ...Line) <-chan Line {
	ch := make(chan Line, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestFanIn_MergesAllLines(t *testing.T) {
	ctx := context.Background()

	ch1 := makeStaticCh(
		Line{Service: "a", Text: "a1"},
		Line{Service: "a", Text: "a2"},
	)
	ch2 := makeStaticCh(
		Line{Service: "b", Text: "b1"},
	)

	out := FanIn(ctx, ch1, ch2)

	var got []string
	for l := range out {
		got = append(got, l.Service+":"+l.Text)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(got), got)
	}

	sort.Strings(got)
	want := []string{"a:a1", "a:a2", "b:b1"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("line[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestFanIn_CancelStopsOutput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Unbuffered channel that will never send — goroutine should unblock via ctx.
	blocking := make(chan Line)
	out := FanIn(ctx, blocking)

	cancel()

	// Drain; channel must close after cancel.
	for range out {
	}
}
