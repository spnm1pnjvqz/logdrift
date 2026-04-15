package filter_test

import (
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/filter"
	"github.com/yourorg/logdrift/internal/runner"
)

func makeLineCh(lines []runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func TestApply_FiltersLines(t *testing.T) {
	f, err := filter.New(filter.Config{Include: []string{"ERROR"}})
	if err != nil {
		t.Fatal(err)
	}

	input := []runner.LogLine{
		{Service: "svc", Line: "ERROR: bad thing"},
		{Service: "svc", Line: "INFO: all good"},
		{Service: "svc", Line: "ERROR: another bad"},
	}

	out := filter.Apply(f, makeLineCh(input))

	var got []runner.LogLine
	for ll := range out {
		got = append(got, ll)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
	for _, ll := range got {
		if ll.Line == "INFO: all good" {
			t.Error("INFO line should have been filtered out")
		}
	}
}

func TestApply_ClosesOutputWhenInputClosed(t *testing.T) {
	f, err := filter.New(filter.Config{})
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan runner.LogLine)
	close(ch)
	out := filter.Apply(f, ch)

	select {
	case _, ok := <-out:
		if ok {
			t.Error("expected output channel to be closed")
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for output channel to close")
	}
}
