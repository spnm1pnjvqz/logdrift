package debounce

import (
	"testing"
	"time"

	"github.com/user/logdrift/internal/runner"
)

func line(svc, text string) runner.LogLine {
	return runner.LogLine{Service: svc, Text: text}
}

func TestNew_ZeroWindow_ReturnsError(t *testing.T) {
	_, err := New(0)
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestNew_NegativeWindow_ReturnsError(t *testing.T) {
	_, err := New(-time.Second)
	if err == nil {
		t.Fatal("expected error for negative window")
	}
}

func TestNew_ValidWindow_NoError(t *testing.T) {
	_, err := New(100 * time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAllow_FirstOccurrence_Passes(t *testing.T) {
	d, _ := New(100 * time.Millisecond)
	if !d.Allow(line("svc", "hello")) {
		t.Error("expected first occurrence to pass")
	}
}

func TestAllow_SecondIdentical_Suppressed(t *testing.T) {
	d, _ := New(200 * time.Millisecond)
	d.Allow(line("svc", "hello"))
	if d.Allow(line("svc", "hello")) {
		t.Error("expected duplicate within window to be suppressed")
	}
}

func TestAllow_DifferentText_Passes(t *testing.T) {
	d, _ := New(200 * time.Millisecond)
	d.Allow(line("svc", "hello"))
	if !d.Allow(line("svc", "world")) {
		t.Error("expected different text to pass")
	}
}

func TestAllow_DifferentService_Independent(t *testing.T) {
	d, _ := New(200 * time.Millisecond)
	d.Allow(line("svc-a", "hello"))
	if !d.Allow(line("svc-b", "hello")) {
		t.Error("expected different service to be independent")
	}
}

func TestAllow_AfterWindowExpires_PassesAgain(t *testing.T) {
	d, _ := New(50 * time.Millisecond)
	d.Allow(line("svc", "hello"))
	time.Sleep(80 * time.Millisecond)
	if !d.Allow(line("svc", "hello")) {
		t.Error("expected line to pass after window expires")
	}
}
