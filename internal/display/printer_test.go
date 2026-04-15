package display

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/differ"
)

func makeEvent(service, text string, isDrift bool) differ.Event {
	return differ.Event{
		Line: differ.Line{
			Service:   service,
			Text:      text,
			Timestamp: time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		},
		IsDrift: isDrift,
	}
}

func TestPrinter_Print_NoDrift(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf)
	p.Print(makeEvent("api", "hello world", false))

	out := buf.String()
	if !strings.Contains(out, "api") {
		t.Errorf("expected service name in output, got: %q", out)
	}
	if !strings.Contains(out, "hello world") {
		t.Errorf("expected log text in output, got: %q", out)
	}
	if strings.Contains(out, "[DRIFT]") {
		t.Errorf("unexpected [DRIFT] tag in non-drift event: %q", out)
	}
}

func TestPrinter_Print_DriftTagPresent(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf)
	p.Print(makeEvent("worker", "something odd", true))

	out := buf.String()
	if !strings.Contains(out, "[DRIFT]") {
		t.Errorf("expected [DRIFT] tag for drift event, got: %q", out)
	}
}

func TestPrinter_ColorStable(t *testing.T) {
	p := New(nil)
	c1 := p.colorFor("svc-a")
	c2 := p.colorFor("svc-a")
	if c1 != c2 {
		t.Error("expected same color for same service name")
	}
}

func TestPrinter_Run_ConsumesAllEvents(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf)

	ch := make(chan differ.Event, 3)
	ch <- makeEvent("a", "line1", false)
	ch <- makeEvent("b", "line2", true)
	ch <- makeEvent("a", "line3", false)
	close(ch)

	p.Run(ch)

	out := buf.String()
	for _, want := range []string{"line1", "line2", "line3", "[DRIFT]"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got:\n%s", want, out)
		}
	}
}
