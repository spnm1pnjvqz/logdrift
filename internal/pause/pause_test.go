package pause_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/pause"
	"github.com/user/logdrift/internal/runner"
)

func makeLineCh(lines []string) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- runner.LogLine{Service: "svc", Text: l}
	}
	close(ch)
	return ch
}

func TestNew_DefaultResumed(t *testing.T) {
	ctrl := pause.New()
	if ctrl.IsPaused() {
		t.Fatal("expected controller to start resumed")
	}
}

func TestPauseResume_TogglesState(t *testing.T) {
	ctrl := pause.New()
	ctrl.Pause()
	if !ctrl.IsPaused() {
		t.Fatal("expected paused after Pause()")
	}
	ctrl.Resume()
	if ctrl.IsPaused() {
		t.Fatal("expected resumed after Resume()")
	}
}

func TestApply_PassesLinesWhenResumed(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ctrl := pause.New()
	in := makeLineCh([]string{"a", "b", "c"})
	out := pause.Apply(ctx, ctrl, in)

	var got []string
	for l := range out {
		got = append(got, l.Text)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(got))
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	ctrl := pause.New()
	blocking := make(chan runner.LogLine) // never closed
	out := pause.Apply(ctx, ctrl, blocking)

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

func TestApply_PauseThenResume_DeliversLines{
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ctrl := pause.New()
	ctrl.Pause()

	in := make(chan runner.LogLine, 1)
	in <- runner.LogLine{Service: "svc", Text: "hello"}
	close(in)

	out := pause.Apply(ctx, ctrl, in)

	// ensure nothing arrives while paused
	select {
	case <-out:
		t.Fatal("received line while paused")
	case <-time.After(50 * time.Millisecond):
	}

	ctrl.Resume()

	select {
	case l, ok := <-out:
		if !ok {
			t.Fatal("channel closed before line arrived")
		}
		if l.Text != "hello" {
			t.Fatalf("unexpected text: %q", l.Text)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for line after resume")
	}
}
