package tee_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/runner"
	"github.com/user/logdrift/internal/tee"
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

func TestTee_BothOutputsReceiveAllLines(t *testing.T) {
	input := []runner.LogLine{
		{Service: "svc", Text: "alpha"},
		{Service: "svc", Text: "beta"},
		{Service: "svc", Text: "gamma"},
	}
	src := makeLineCh(input)
	ctx := context.Background()

	a, b := tee.Tee(ctx, src)

	var got1, got2 []runner.LogLine
	done := make(chan struct{})
	go func() {
		got1 = collect(a)
		close(done)
	}()
	got2 = collect(b)
	<-done

	if len(got1) != len(input) {
		t.Fatalf("out1: want %d lines, got %d", len(input), len(got1))
	}
	if len(got2) != len(input) {
		t.Fatalf("out2: want %d lines, got %d", len(input), len(got2))
	}
	for i, l := range input {
		if got1[i].Text != l.Text {
			t.Errorf("out1[%d]: want %q, got %q", i, l.Text, got1[i].Text)
		}
		if got2[i].Text != l.Text {
			t.Errorf("out2[%d]: want %q, got %q", i, l.Text, got2[i].Text)
		}
	}
}

func TestTee_ChannelsClosedWhenSourceClosed(t *testing.T) {
	src := makeLineCh(nil)
	ctx := context.Background()
	a, b := tee.Tee(ctx, src)

	select {
	case _, ok := <-a:
		if ok {
			t.Fatal("out1 should be closed")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for out1 to close")
	}
	select {
	case _, ok := <-b:
		if ok {
			t.Fatal("out2 should be closed")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for out2 to close")
	}
}

func TestTee_CancelStopsOutput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Unbuffered infinite source
	src := make(chan runner.LogLine)
	a, b := tee.Tee(ctx, src)

	cancel()

	// Drain both; they should close promptly after cancel.
	timeout := time.After(time.Second)
	for i, ch := range []<-chan runner.LogLine{a, b} {
		select {
		case <-ch:
		case <-timeout:
			t.Fatalf("out%d did not close after cancel", i+1)
		}
	}
}
