package multiline_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/multiline"
	"github.com/user/logdrift/internal/runner"
)

func feed(lines []string, service string) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- runner.LogLine{Service: service, Text: l}
	}
	close(ch)
	return ch
}

func collect(ctx context.Context, ch <-chan runner.LogLine) []runner.LogLine {
	var out []runner.LogLine
	for l := range ch {
		out = append(out, l)
	}
	return out
}

func TestNew_InvalidPattern_ReturnsError(t *testing.T) {
	_, err := multiline.New("[", time.Second)
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}

func TestNew_ValidPattern_NoError(t *testing.T) {
	_, err := multiline.New(`^\d{4}-`, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_CoalescesLines(t *testing.T) {
	j, _ := multiline.New(`^START`, 200*time.Millisecond)
	input := []string{
		"START first",
		"  continuation 1",
		"  continuation 2",
		"START second",
		"  cont",
	}
	ctx := context.Background()
	result := collect(ctx, j.Apply(ctx, feed(input, "svc")))
	if len(result) != 2 {
		t.Fatalf("expected 2 logical lines, got %d", len(result))
	}
	if result[0].Service != "svc" {
		t.Errorf("expected service svc, got %s", result[0].Service)
	}
}

func TestApply_SingleLine_PassedThrough(t *testing.T) {
	j, _ := multiline.New(`^START`, 200*time.Millisecond)
	input := []string{"START only"}
	ctx := context.Background()
	result := collect(ctx, j.Apply(ctx, feed(input, "svc")))
	if len(result) != 1 {
		t.Fatalf("expected 1 line, got %d", len(result))
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	j, _ := multiline.New(`^\d`, 2*time.Second)
	ch := make(chan runner.LogLine)
	ctx, cancel := context.WithCancel(context.Background())
	out := j.Apply(ctx, ch)
	cancel()
	for range out {
	}
}
