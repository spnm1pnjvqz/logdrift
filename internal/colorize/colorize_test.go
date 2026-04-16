package colorize

import (
	"strings"
	"testing"
)

func TestNew_EmptyAssignments(t *testing.T) {
	c := New()
	if len(c.assigned) != 0 {
		t.Fatalf("expected empty assignments, got %d", len(c.assigned))
	}
}

func TestWrap_ContainsANSI(t *testing.T) {
	c := New()
	out := c.Wrap("svc", "hello")
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("expected ANSI escape in output, got %q", out)
	}
	if !strings.Contains(out, "hello") {
		t.Fatalf("original text missing from output")
	}
	if !strings.HasSuffix(out, "\x1b[0m") {
		t.Fatalf("expected reset suffix")
	}
}

func TestWrap_SameServiceSameColor(t *testing.T) {
	c := New()
	a := c.Wrap("alpha", "x")
	b := c.Wrap("alpha", "x")
	if a != b {
		t.Fatalf("same service produced different colors: %q vs %q", a, b)
	}
}

func TestWrap_DifferentServicesDistinctColors(t *testing.T) {
	c := New()
	a := c.ServiceColor("svc-a")
	b := c.ServiceColor("svc-b")
	if a == b {
		t.Fatalf("expected distinct colors for different services")
	}
}

func TestWrap_PaletteWrapsAround(t *testing.T) {
	c := New()
	services := []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8"}
	for _, s := range services {
		out := c.Wrap(s, "msg")
		if !strings.Contains(out, "msg") {
			t.Fatalf("text missing for service %s", s)
		}
	}
	// s8 wraps around; ensure it still produces valid output
	out := c.Wrap("s8", "msg")
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("expected ANSI escape after palette wrap")
	}
}

func TestServiceColor_Format(t *testing.T) {
	c := New()
	color := c.ServiceColor("mysvc")
	if !strings.HasPrefix(color, "\x1b[") {
		t.Fatalf("expected ANSI prefix, got %q", color)
	}
	if !strings.HasSuffix(color, "m") {
		t.Fatalf("expected 'm' suffix, got %q", color)
	}
}
