package batch

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

func TestNew_InvalidSize_ReturnsError(t *testing.T) {
	_, err := New(0, 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected error for size=0")
	}
}

func TestNew_InvalidInterval_ReturnsError(t *testing.T) {
	_, err := New(5, 0)
	if err == nil {
		t.Fatal("expected error for interval=0")
	}
}

func TestNew_ValidArgs_NoError(t *testing.T) {
	_, err := New(5, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_BatchesBySize(t *testing.T) {
	lines := []runner.LogLine{
		{Service: "svc", Text: "a"},
		{Service: "svc", Text: "b"},
		{Service: "svc", Text: "c"},
		{Service: "svc", Text: "d"},
	}
	b, _ := New(2, 500*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	out := b.Apply(ctx, makeLineCh(lines))
	var batches [][]runner.LogLine
	for batch := range out {
		batches = append(batches, batch)
	}
	if len(batches) != 2 {
		t.Fatalf("expected 2 batches, got %d", len(batches))
	}
	if len(batches[0]) != 2 || len(batches[1]) != 2 {
		t.Fatalf("unexpected batch sizes: %v", batches)
	}
}

func TestApply_FlushesRemainder(t *testing.T) {
	lines := []runner.LogLine{
		{Service: "svc", Text: "x"},
		{Service: "svc", Text: "y"},
		{Service: "svc", Text: "z"},
	}
	b, _ := New(10, 500*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	out := b.Apply(ctx, makeLineCh(lines))
	var batches [][]runner.LogLine
	for batch := range out {
		batches = append(batches, batch)
	}
	if len(batches) != 1 || len(batches[0]) != 3 {
		t.Fatalf("expected 1 batch of 3, got %v", batches)
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	ch := make(chan runner.LogLine) // never sends
	b, _ := New(5, 50*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	out := b.Apply(ctx, ch)
	cancel()
	_, ok := <-out
	if ok {
		t.Fatal("expected channel to be closed after cancel")
	}
}
