package lineformat

import (
	"context"
	"strings"
	"testing"

	"github.com/user/logdrift/internal/runner"
)

func makeLine(service, text string) runner.LogLine {
	return runner.LogLine{Service: service, Text: text}
}

func makeLineCh(lines ...runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestNew_EmptyTemplate_ReturnsError(t *testing.T) {
	_, err := New("")
	if err == nil {
		t.Fatal("expected error for empty template")
	}
}

func TestNew_NoPlaceholder_ReturnsError(t *testing.T) {
	_, err := New("hello world")
	if err == nil {
		t.Fatal("expected error for template with no placeholder")
	}
}

func TestNew_ValidTemplate_NoError(t *testing.T) {
	_, err := New("[{service}] {text}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFormat_ServiceAndText(t *testing.T) {
	f, _ := New("[{service}] {text}")
	result := f.Format(makeLine("api", "started"))
	if !strings.Contains(result, "[api]") || !strings.Contains(result, "started") {
		t.Errorf("unexpected format output: %q", result)
	}
}

func TestFormat_TimePresent(t *testing.T) {
	f, _ := New("{time} {text}")
	result := f.Format(makeLine("svc", "msg"))
	if !strings.Contains(result, "msg") {
		t.Errorf("text missing from output: %q", result)
	}
	// {time} should have been replaced (not literally present)
	if strings.Contains(result, "{time}") {
		t.Errorf("{time} placeholder not replaced: %q", result)
	}
}

func TestApply_FormatsAllLines(t *testing.T) {
	f, _ := New(">> [{service}] {text}")
	ch := makeLineCh(
		makeLine("web", "req1"),
		makeLine("db", "query"),
	)
	ctx := context.Background()
	out := Apply(ctx, f, ch)
	var results []runner.LogLine
	for l := range out {
		results = append(results, l)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(results))
	}
	if !strings.HasPrefix(results[0].Text, ">> [web]") {
		t.Errorf("unexpected line: %q", results[0].Text)
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	f, _ := New("{service}: {text}")
	blocking := make(chan runner.LogLine)
	ctx, cancel := context.WithCancel(context.Background())
	out := Apply(ctx, f, blocking)
	cancel()
	_, ok := <-out
	if ok {
		t.Error("expected channel to be closed after cancel")
	}
}
