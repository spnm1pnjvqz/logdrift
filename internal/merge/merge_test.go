package merge_test

import (
	"context"
	"testing"
	"time"

	"github.com/logdrift/logdrift/internal/merge"
	"github.com/logdrift/logdrift/internal/runner"
)

func makeCh(lines ...string) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- runner.LogLine{Service: "svc", Text: l}
	}
	close(ch)
	return ch
}

func TestMerge_NoSources_ReturnsError(t *testing.T) {
	_, err := merge.Merge(context.Background())
	if err == nil {
		t.Fatal("expected error for zero sources")
	}
}

func TestMerge_SingleSource_EmitsAllLines(t *testing.T) {
	ch := makeCh("a", "b", "c")
	out, err := merge.Merge(context.Background(), ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got []string
	for line := range out {
		got = append(got, line.Text)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(got))
	}
}

func TestMerge_MultipleSources_EmitsAll(t *testing.T) {
	ch1 := makeCh("x", "y")
	ch2 := makeCh("p", "q", "r")
	out, err := merge.Merge(context.Background(), ch1, ch2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count := 0
	for range out {
		count++
	}
	if count != 5 {
		t.Fatalf("expected 5 lines, got %d", count)
	}
}

func TestMerge_Cancel_StopsOutput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	blocking := make(chan runner.LogLine)
	out, err := merge.Merge(ctx, blocking)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to close after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for channel close")
	}
}
