package label

import (
	"context"
	"testing"
	"time"

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

func TestNew_EmptyService_ReturnsError(t *testing.T) {
	_, err := New("")
	if err == nil {
		t.Fatal("expected error for empty service name")
	}
}

func TestNew_ValidService_NoError(t *testing.T) {
	_, err := New("svc-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_StampsService(t *testing.T) {
	lbr, _ := New("api")
	input := []runner.LogLine{
		{Service: "old", Text: "hello"},
		{Service: "", Text: "world"},
	}
	ctx := context.Background()
	out := lblr.Apply(ctx, makeLineCh(input))

	var got []runner.LogLine
	for l := range out {
		got = append(got, l)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
	for _, l := range got {
		if l.Service != "api" {
			t.Errorf("expected service=api, got %q", l.Service)
		}
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	lbr, _ := New("svc")
	blocking := make(chan runner.LogLine)
	ctx, cancel := context.WithCancel(context.Background())
	out := lblr.Apply(ctx, blocking)
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
