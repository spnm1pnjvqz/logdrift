package since_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/runner"
	"github.com/user/logdrift/internal/since"
)

// TestApply_OnlyRecentLinesEmitted sends a mix of old and new lines through
// the filter and asserts ordering and count are preserved correctly.
func TestApply_OnlyRecentLinesEmitted(t *testing.T) {
	cutoff := time.Date(2024, 3, 15, 9, 0, 0, 0, time.UTC)
	f, err := since.New(cutoff, []string{"2006-01-02 15:04:05"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	lines := []runner.LogLine{
		{Service: "api", Text: "2024-03-15 08:59:59 startup"},
		{Service: "api", Text: "2024-03-15 09:00:00 ready"},
		{Service: "api", Text: "2024-03-15 09:01:00 request received"},
		{Service: "api", Text: "2024-03-14 22:00:00 yesterday"},
		{Service: "api", Text: "unparseable log line"},
	}

	in := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		in <- l
	}
	close(in)

	ctx := context.Background()
	var got []runner.LogLine
	for l := range f.Apply(ctx, in) {
		got = append(got, l)
	}

	// expect: 09:00:00, 09:01:00, unparseable
	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(got), got)
	}
	if got[0].Text != "2024-03-15 09:00:00 ready" {
		t.Errorf("unexpected first line: %q", got[0].Text)
	}
	if got[2].Text != "unparseable log line" {
		t.Errorf("expected unparseable line last, got %q", got[2].Text)
	}
}
