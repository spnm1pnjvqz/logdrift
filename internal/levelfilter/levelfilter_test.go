package levelfilter_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/levelfilter"
	"github.com/yourorg/logdrift/internal/runner"
)

func TestNew_UnknownLevel_ReturnsError(t *testing.T) {
	_, err := levelfilter.New("verbose")
	if err == nil {
		t.Fatal("expected error for unknown level")
	}
}

func TestNew_ValidLevel_NoError(t *testing.T) {
	for _, lvl := range []string{"debug", "info", "warn", "error", "INFO", "WARN"} {
		_, err := levelfilter.New(lvl)
		if err != nil {
			t.Fatalf("unexpected error for level %q: %v", lvl, err)
		}
	}
}

func TestAllow_MinInfo_DropsDebug(t *testing.T) {
	f, _ := levelfilter.New("info")
	if f.Allow("some debug message") {
		t.Error("expected debug line to be dropped at info threshold")
	}
}

func TestAllow_MinInfo_PassesInfo(t *testing.T) {
	f, _ := levelfilter.New("info")
	if !f.Allow("INFO starting server") {
		t.Error("expected info line to pass")
	}
}

func TestAllow_MinWarn_DropsInfo(t *testing.T) {
	f, _ := levelfilter.New("warn")
	if f.Allow("INFO all good") {
		t.Error("expected info line to be dropped at warn threshold")
	}
}

func TestAllow_MinError_PassesError(t *testing.T) {
	f, _ := levelfilter.New("error")
	if !f.Allow("ERROR something failed") {
		t.Error("expected error line to pass")
	}
}

func makeLineCh(lines []string) <-chan runner.LogLine {
	ch := make(chan runner.LogLine, len(lines))
	for _, l := range lines {
		ch <- runner.LogLine{Service: "svc", Text: l}
	}
	close(ch)
	return ch
}

func TestApply_FiltersLines(t *testing.T) {
	f, _ := levelfilter.New("warn")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	in := makeLineCh([]string{
		"debug noise",
		"INFO startup complete",
		"WARN disk usage high",
		"ERROR disk full",
	})

	out := f.Apply(ctx, in)
	var got []string
	for line := range out {
		got = append(got, line.Text)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(got), got)
	}
}

func TestApply_Cancel_StopsOutput(t *testing.T) {
	f, _ := levelfilter.New("debug")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	in := make(chan runner.LogLine)
	out := f.Apply(ctx, in)
	select {
	case _, ok := <-out:
		if ok {
			t.Error("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for channel close")
	}
}
