package since_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/runner"
	"github.com/user/logdrift/internal/since"
)

func makeLineCh(texts []string) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(texts))
	for _, t := range texts {
		ch <- runner.LogLine{Service: "svc", Text: t}
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

func TestNew_ZeroCutoff_ReturnsError(t *testing.T) {
	_, err := since.New(time.Time{}, nil)
	if err == nil {
		t.Fatal("expected error for zero cutoff")
	}
}

func TestNew_ValidCutoff_NoError(t *testing.T) {
	_, err := since.New(time.Now(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAllow_UnparsableLine_PassesThrough(t *testing.T) {
	f, _ := since.New(time.Now(), nil)
	line := runner.LogLine{Text: "no timestamp here"}
	if !f.Allow(line) {
		t.Error("expected unparsable line to pass through")
	}
}

func TestAllow_OldTimestamp_Dropped(t *testing.T) {
	cutoff := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	f, _ := since.New(cutoff, []string{time.RFC3339})
	old := runner.LogLine{Text: "2024-01-09T23:59:59Z rest of message"}
	if f.Allow(old) {
		t.Error("expected old line to be dropped")
	}
}

func TestAllow_ExactCutoff_Passes(t *testing.T) {
	cutoff := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	f, _ := since.New(cutoff, []string{time.RFC3339})
	exact := runner.LogLine{Text: "2024-01-10T00:00:00Z rest of message"}
	if !f.Allow(exact) {
		t.Error("expected exact cutoff line to pass")
	}
}

func TestApply_FiltersOldLines(t *testing.T) {
	cutoff := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	f, _ := since.New(cutoff, []string{time.RFC3339})

	input := []string{
		"2024-06-01T11:59:59Z old line",
		"2024-06-01T12:00:00Z exact",
		"2024-06-01T13:00:00Z newer",
		"no timestamp",
	}
	ctx := context.Background()
	result := collect(f.Apply(ctx, makeLineCh(input)))
	if len(result) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(result))
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	f, _ := since.New(time.Now().Add(-time.Hour), nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ch := make(chan runner.LogLine)
	result := collect(f.Apply(ctx, ch))
	if len(result) != 0 {
		t.Errorf("expected no output after cancel, got %d", len(result))
	}
}
