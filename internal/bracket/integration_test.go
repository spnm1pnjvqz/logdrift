package bracket_test

import (
	"context"
	"testing"

	"github.com/user/logdrift/internal/bracket"
	"github.com/user/logdrift/internal/runner"
)

func sendLines(lines []runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestIntegration_BracketPipeline(t *testing.T) {
	// Chain two bracketing stages: first adds parens, then adds angle brackets.
	inner, err := bracket.New("(", ")")
	if err != nil {
		t.Fatal(err)
	}
	outer, err := bracket.New("<", ">")
	if err != nil {
		t.Fatal(err)
	}

	src := sendLines([]runner.LogLine{
		{Service: "svc", Text: "hello"},
		{Service: "svc", Text: "world"},
	})

	ctx := context.Background()
	stage1 := inner.Apply(ctx, src)
	stage2 := outer.Apply(ctx, stage1)

	var results []string
	for l := range stage2 {
		results = append(results, l.Text)
	}

	expected := []string{"<(hello)>", "<(world)>"}
	if len(results) != len(expected) {
		t.Fatalf("got %d results, want %d", len(results), len(expected))
	}
	for i, r := range results {
		if r != expected[i] {
			t.Errorf("[%d] got %q, want %q", i, r, expected[i])
		}
	}
}
