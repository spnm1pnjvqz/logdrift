package gate_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/gate"
	"github.com/yourorg/logdrift/internal/runner"
)

func makeLogLineCh(lines []runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collectGateLines(ctx context.Context, ch <-chan runner.LogLine) []runner.LogLine {
	var out []runner.LogLine
	for {
		select {
		case l, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, l)
		case <-ctx.Done():
			return out
		}
	}
}

func TestNew_NilCondition_ReturnsError(t *testing.T) {
	_, err := gate.New(nil)
	if err == nil {
		t.Fatal("expected error for nil condition, got nil")
	}
}

func TestNew_ValidCondition_NoError(t *testing.T) {
	_, err := gate.New(func(l runner.LogLine) bool { return true })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_PassesMatchingLines(t *testing.T) {
	g, err := gate.New(func(l runner.LogLine) bool {
		return l.Service == "web"
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := []runner.LogLine{
		{Service: "web", Text: "request received"},
		{Service: "db", Text: "query executed"},
		{Service: "web", Text: "response sent"},
	}

	ctx := context.Background()
	out := g.Apply(ctx, makeLogLineCh(lines))
	got := collectGateLines(ctx, out)

	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
	for _, l := range got {
		if l.Service != "web" {
			t.Errorf("expected service=web, got %q", l.Service)
		}
	}
}

func TestApply_DropsNonMatchingLines(t *testing.T) {
	g, err := gate.New(func(l runner.LogLine) bool {
		return false
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := []runner.LogLine{
		{Service: "web", Text: "hello"},
		{Service: "db", Text: "world"},
	}

	ctx := context.Background()
	out := g.Apply(ctx, makeLogLineCh(lines))
	got := collectGateLines(ctx, out)

	if len(got) != 0 {
		t.Fatalf("expected 0 lines, got %d", len(got))
	}
}

func TestApply_ClosesWhenInputClosed(t *testing.T) {
	g, err := gate.New(func(l runner.LogLine) bool { return true })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ch := make(chan runner.LogLine)
	close(ch)

	ctx := context.Background()
	out := g.Apply(ctx, ch)

	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for channel close")
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	g, err := gate.New(func(l runner.LogLine) bool { return true })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	src := make(chan runner.LogLine)
	ctx, cancel := context.WithCancel(context.Background())
	out := g.Apply(ctx, src)

	cancel()

	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for channel close after cancel")
	}
}
