package mask_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/mask"
	"github.com/yourorg/logdrift/internal/runner"
)

func TestNew_NoPatterns_ReturnsError(t *testing.T) {
	_, err := mask.New(nil, "")
	if err == nil {
		t.Fatal("expected error for empty patterns")
	}
}

func TestNew_InvalidPattern_ReturnsError(t *testing.T) {
	_, err := mask.New([]string{"[invalid"}, "")
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestNew_ValidPatterns_NoError(t *testing.T) {
	_, err := mask.New([]string{`\d+`}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApply_DefaultPlaceholder(t *testing.T) {
	m, _ := mask.New([]string{`password=\S+`}, "")
	got := m.Apply("user login password=secret123")
	want := "user login [MASKED]"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestApply_CustomPlaceholder(t *testing.T) {
	m, _ := mask.New([]string{`\d{4}-\d{4}-\d{4}-\d{4}`}, "****")
	got := m.Apply("card: 1234-5678-9012-3456")
	if got != "card: ****" {
		t.Errorf("unexpected result: %q", got)
	}
}

func TestApply_MultipleRules(t *testing.T) {
	m, _ := mask.New([]string{`token=\S+`, `secret=\S+`}, "")
	got := m.Apply("token=abc secret=xyz other=ok")
	want := "[MASKED] [MASKED] other=ok"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
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

func TestTransform_MasksText(t *testing.T) .New([]string{`\d+`}, "NUM")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	in := makeLineCh([]runner.LogLine{
		{Service: "svc", Text: "got 42 errors"},
		{Service: "svc", Text: "no digits here"},
	})
	out := m.Transform(ctx, in)

	var results []string
	for l := range out {
		results = append(results, l.Text)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(results))
	}
	if results[0] != "got NUM errors" {
		t.Errorf("unexpected: %q", results[0])
	}
	if results[1] != "no digits here" {
		t.Errorf("unexpected: %q", results[1])
	}
}

func TestTransform_Cancel_StopsOutput(t *testing.T) {
	m, _ := mask.New([]string{`.`}, "")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	blocked := make(chan runner.LogLine)
	out := m.Transform(ctx, blocked)
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected closed channel")
		}
	case <-time.After(time.Second):
		t.Fatal("channel not closed after cancel")
	}
}
