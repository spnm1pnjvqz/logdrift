package buffer

import (
	"fmt"
	"testing"

	"github.com/user/logdrift/internal/runner"
)

func makeLine(svc, text string) runner.LogLine {
	return runner.LogLine{Service: svc, Text: text}
}

func TestNew_NegativeCapacity_ReturnsError(t *testing.T) {
	_, err := New(-1)
	if err == nil {
		t.Fatal("expected error for negative capacity")
	}
}

func TestNew_ZeroCapacity_UsesDefault(t *testing.T) {
	b, err := New(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.cap != defaultCapacity {
		t.Fatalf("expected cap %d, got %d", defaultCapacity, b.cap)
	}
}

func TestAdd_StoresLine(t *testing.T) {
	b, _ := New(10)
	b.Add(makeLine("svcA", "hello"))

	lines := b.Get("svcA")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0].Text != "hello" {
		t.Errorf("expected 'hello', got %q", lines[0].Text)
	}
}

func TestAdd_EvictsOldestWhenFull(t *testing.T) {
	cap := 3
	b, _ := New(cap)

	for i := 0; i < 5; i++ {
		b.Add(makeLine("svc", fmt.Sprintf("line%d", i)))
	}

	lines := b.Get("svc")
	if len(lines) != cap {
		t.Fatalf("expected %d lines, got %d", cap, len(lines))
	}
	if lines[0].Text != "line2" {
		t.Errorf("expected oldest retained line to be 'line2', got %q", lines[0].Text)
	}
	if lines[2].Text != "line4" {
		t.Errorf("expected newest line to be 'line4', got %q", lines[2].Text)
	}
}

func TestGet_UnknownService_ReturnsNil(t *testing.T) {
	b, _ := New(10)
	if got := b.Get("missing"); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestServices_ReturnsAllServices(t *testing.T) {
	b, _ := New(10)
	b.Add(makeLine("alpha", "a"))
	b.Add(makeLine("beta", "b"))

	svcs := b.Services()
	if len(svcs) != 2 {
		t.Fatalf("expected 2 services, got %d", len(svcs))
	}
}

func TestLen_ReturnsCorrectCount(t *testing.T) {
	b, _ := New(10)
	b.Add(makeLine("svc", "x"))
	b.Add(makeLine("svc", "y"))

	if n := b.Len("svc"); n != 2 {
		t.Errorf("expected 2, got %d", n)
	}
	if n := b.Len("other"); n != 0 {
		t.Errorf("expected 0 for unknown service, got %d", n)
	}
}
