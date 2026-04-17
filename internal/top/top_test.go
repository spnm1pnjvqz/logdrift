package top

import (
	"context"
	"testing"

	"github.com/user/logdrift/internal/runner"
)

func TestNew_InvalidN_ReturnsError(t *testing.T) {
	_, err := New(0)
	if err == nil {
		t.Fatal("expected error for n=0")
	}
}

func TestNew_ValidN_NoError(t *testing.T) {
	_, err := New(3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTop_ReturnsTopN(t *testing.T) {
	tr, _ := New(2)
	for i := 0; i < 5; i++ {
		tr.Add("svc", "alpha")
	}
	for i := 0; i < 3; i++ {
		tr.Add("svc", "beta")
	}
	tr.Add("svc", "gamma")

	entries := tr.Top("svc")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Line != "alpha" || entries[0].Count != 5 {
		t.Errorf("unexpected top entry: %+v", entries[0])
	}
	if entries[1].Line != "beta" || entries[1].Count != 3 {
		t.Errorf("unexpected second entry: %+v", entries[1])
	}
}

func TestTop_UnknownService_ReturnsEmpty(t *testing.T) {
	tr, _ := New(5)
	if got := tr.Top("missing"); len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

func TestTop_IndependentPerService(t *testing.T) {
	tr, _ := New(1)
	tr.Add("a", "foo")
	tr.Add("a", "foo")
	tr.Add("b", "bar")

	if tr.Top("a")[0].Line != "foo" {
		t.Error("expected foo for service a")
	}
	if tr.Top("b")[0].Line != "bar" {
		t.Error("expected bar for service b")
	}
}

func makeLineCh(lines []runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestApply_ForwardsLines(t *testing.T) {
	tr, _ := New(3)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input := []runner.LogLine{
		{Service: "svc", Text: "hello"},
		{Service: "svc", Text: "world"},
		{Service: "svc", Text: "hello"},
	}
	out := tr.Apply(ctx, makeLineCh(input))
	var collected []runner.LogLine
	for l := range out {
		collected = append(collected, l)
	}
	if len(collected) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(collected))
	}
	if tr.Top("svc")[0].Line != "hello" {
		t.Error("expected hello as top line")
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	tr, _ := New(2)
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan runner.LogLine)
	out := tr.Apply(ctx, ch)
	cancel()
	_, ok := <-out
	if ok {
		t.Error("expected channel to be closed after cancel")
	}
}
