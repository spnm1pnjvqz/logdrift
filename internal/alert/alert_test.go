package alert

import (
	"testing"
)

func TestNew_NoPatterns_ReturnsError(t *testing.T) {
	_, err := New(map[string]string{})
	if err == nil {
		t.Fatal("expected error for empty patterns")
	}
}

func TestNew_InvalidPattern_ReturnsError(t *testing.T) {
	_, err := New(map[string]string{"bad": "[invalid"})
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestNew_ValidPatterns_NoError(t *testing.T) {
	_, err := New(map[string]string{"err": "ERROR", "warn": "WARN"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheck_NoMatch_ReturnsEmpty(t *testing.T) {
	a, _ := New(map[string]string{"err": "ERROR"})
	evs := a.Check("svc", "everything is fine")
	if len(evs) != 0 {
		t.Fatalf("expected no events, got %d", len(evs))
	}
}

func TestCheck_Match_ReturnsEvent(t *testing.T) {
	a, _ := New(map[string]string{"err": "ERROR"})
	evs := a.Check("api", "ERROR: something broke")
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Rule != "err" || evs[0].Service != "api" {
		t.Fatalf("unexpected event: %+v", evs[0])
	}
}

func TestCheck_MultipleRules_AllMatch(t *testing.T) {
	a, _ := New(map[string]string{"err": "ERROR", "panic": "panic"})
	evs := a.Check("svc", "ERROR panic occurred")
	if len(evs) != 2 {
		t.Fatalf("expected 2 events, got %d", len(evs))
	}
}
