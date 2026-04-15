package dedupe_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/dedupe"
	"github.com/user/logdrift/internal/runner"
)

func makeLineCh(lines []runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collect(ch <-chan runner.LogLine) []runner.LogLine {
	var out []runner.LogLine
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestIsDuplicate_FirstOccurrence_NotDuplicate(t *testing.T) {
	d := dedupe.New()
	if d.IsDuplicate("svc", "hello") {
		t.Fatal("first occurrence should not be a duplicate")
	}
}

func TestIsDuplicate_SameLine_IsDuplicate(t *testing.T) {
	d := dedupe.New()
	d.IsDuplicate("svc", "hello")
	if !d.IsDuplicate("svc", "hello") {
		t.Fatal("repeated line should be a duplicate")
	}
}

func TestIsDuplicate_DifferentServices_Independent(t *testing.T) {
	d := dedupe.New()
	d.IsDuplicate("svc-a", "hello")
	if d.IsDuplicate("svc-b", "hello") {
		t.Fatal("same line on different service should not be a duplicate")
	}
}

func TestReset_ClearsHistory(t *testing.T) {
	d := dedupe.New()
	d.IsDuplicate("svc", "hello")
	d.Reset()
	if d.IsDuplicate("svc", "hello") {
		t.Fatal("after reset the line should not be considered a duplicate")
	}
}

func TestApply_DropsConsecutiveDuplicates(t *testing.T) {
	lines := []runner.LogLine{
		{Service: "svc", Line: "aaa"},
		{Service: "svc", Line: "aaa"}, // duplicate
		{Service: "svc", Line: "bbb"},
		{Service: "svc", Line: "bbb"}, // duplicate
		{Service: "svc", Line: "aaa"}, // not duplicate (last was bbb)
	}
	d := dedupe.New()
	got := collect(dedupe.Apply(context.Background(), d, makeLineCh(lines)))
	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(got))
	}
}

func TestApply_CancelStopsOutput(t *testing.T) {
	ch := make(chan runner.LogLine) // never sends
	ctx, cancel := context.WithCancel(context.Background())
	out := dedupe.Apply(ctx, dedupe.New(), ch)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("channel not closed after cancel")
	}
}
