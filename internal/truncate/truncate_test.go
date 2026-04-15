package truncate

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/user/logdrift/internal/runner"
)

func TestNew_NegativeMaxBytes_ReturnsError(t *testing.T) {
	_, err := New(-1, "...")
	if err == nil {
		t.Fatal("expected error for negative maxBytes")
	}
}

func TestNew_ZeroMaxBytes_UsesDefault(t *testing.T) {
	tr, err := New(0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.maxBytes != defaultMaxBytes {
		t.Errorf("expected %d, got %d", defaultMaxBytes, tr.maxBytes)
	}
}

func TestTruncate_ShortLine_Unchanged(t *testing.T) {
	tr, _ := New(10, "...")
	input := "hello"
	if got := tr.Truncate(input); got != input {
		t.Errorf("expected %q, got %q", input, got)
	}
}

func TestTruncate_ExactLimit_Unchanged(t *testing.T) {
	tr, _ := New(5, "...")
	input := "hello"
	if got := tr.Truncate(input); got != input {
		t.Errorf("expected %q, got %q", input, got)
	}
}

func TestTruncate_LongLine_ClippedWithSuffix(t *testing.T) {
	tr, _ := New(5, "...")
	input := "hello world"
	want := "hello..."
	if got := tr.Truncate(input); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestTruncate_DefaultSuffix(t *testing.T) {
	tr, _ := New(4, "")
	result := tr.Truncate("abcdefgh")
	if !strings.HasSuffix(result, "...") {
		t.Errorf("expected default suffix '...', got %q", result)
	}
}

func makeLineCh(lines []string) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- runner.LogLine{Service: "svc", Line: l}
	}
	close(ch)
	return ch
}

func TestApply_TruncatesLongLines(t *testing.T) {
	tr, _ := New(5, "...")
	in := makeLineCh([]string{"short", "this is too long"})
	ctx := context.Background()
	out := tr.Apply(ctx, in)

	var results []runner.LogLine
	for ll := range out {
		results = append(results, ll)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(results))
	}
	if results[0].Line != "short" {
		t.Errorf("expected 'short', got %q", results[0].Line)
	}
	if results[1].Line != "this ..." {
		t.Errorf("expected 'this ...', got %q", results[1].Line)
	}
}

func TestApply_CancelStopsOutput(t *testing.T) {
	tr, _ := New(10, "...")
	ch := make(chan runner.LogLine)
	ctx, cancel := context.WithCancel(context.Background())
	out := tr.Apply(ctx, ch)
	cancel()

	select {
	case _, ok := <-out:
		if ok {
			t.Error("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for channel close")
	}
}
