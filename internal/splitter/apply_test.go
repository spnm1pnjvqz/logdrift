package splitter

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/runner"
)

func makeLineCh(lines []runner.LogLine) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- l
	}
	close(ch)
	return ch
}

func collect(ch <-chan runner.LogLine, timeout time.Duration) []runner.LogLine {
	var out []runner.LogLine
	timer := time.After(timeout)
	for {
		select {
		case l, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, l)
		case <-timer:
			return out
		}
	}
}

func TestApply_RoutesLinesToCorrectBuckets(t *testing.T) {
	s, _ := New(map[string]string{
		"errors": `(?i)error`,
		"info":   `(?i)info`,
	}, "other")

	src := makeLineCh([]runner.LogLine{
		makeLine("svc", "ERROR: boom"),
		makeLine("svc", "INFO: started"),
		makeLine("svc", "debug: verbose"),
	})

	outputs := MakeOutputs([]string{"errors", "info", "other"}, 8)
	s.Apply(context.Background(), src, outputs)

	errors := collect(outputs["errors"], time.Second)
	infos := collect(outputs["info"], time.Second)
	others := collect(outputs["other"], time.Second)

	if len(errors) != 1 {
		t.Errorf("errors bucket: want 1 line, got %d", len(errors))
	}
	if len(infos) != 1 {
		t.Errorf("info bucket: want 1 line, got %d", len(infos))
	}
	if len(others) != 1 {
		t.Errorf("other bucket: want 1 line, got %d", len(others))
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	s, _ := New(map[string]string{"errors": `error`}, "")

	src := make(chan runner.LogLine) // never sends
	outputs := MakeOutputs([]string{"errors"}, 4)

	ctx, cancel := context.WithCancel(context.Background())
	s.Apply(ctx, src, outputs)
	cancel()

	select {
	case _, ok := <-outputs["errors"]:
		if ok {
			t.Error("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for channel close")
	}
}

func TestApply_UnmatchedBucketDropped(t *testing.T) {
	s, _ := New(map[string]string{"errors": `error`}, "other")

	// outputs map does NOT include "other", so unmatched lines should be dropped
	src := makeLineCh([]runner.LogLine{
		makeLine("svc", "info: harmless"),
		makeLine("svc", "error: bad"),
	})

	outputs := MakeOutputs([]string{"errors"}, 8)
	s.Apply(context.Background(), src, outputs)

	got := collect(outputs["errors"], time.Second)
	if len(got) != 1 {
		t.Errorf("want 1 error line, got %d", len(got))
	}
}
