package top_test

import (
	"context"
	"testing"

	"github.com/user/logdrift/internal/runner"
	"github.com/user/logdrift/internal/top"
)

func TestApply_TopNAccumulatesAcrossStream(t *testing.T) {
	tr, err := top.New(2)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	lines := []runner.LogLine{
		{Service: "web", Text: "GET /health"},
		{Service: "web", Text: "GET /health"},
		{Service: "web", Text: "GET /health"},
		{Service: "web", Text: "POST /login"},
		{Service: "web", Text: "POST /login"},
		{Service: "web", Text: "DELETE /item"},
	}

	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)

	ctx := context.Background()
	out := tr.Apply(ctx, ch)
	var count int
	for range out {
		count++
	}
	if count != len(lines) {
		t.Fatalf("expected %d forwarded lines, got %d", len(lines), count)
	}

	entries := tr.Top("web")
	if len(entries) != 2 {
		t.Fatalf("expected top 2, got %d", len(entries))
	}
	if entries[0].Line != "GET /health" {
		t.Errorf("expected GET /health as #1, got %q", entries[0].Line)
	}
	if entries[0].Count != 3 {
		t.Errorf("expected count 3, got %d", entries[0].Count)
	}
	if entries[1].Line != "POST /login" {
		t.Errorf("expected POST /login as #2, got %q", entries[1].Line)
	}
}
