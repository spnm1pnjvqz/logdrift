package window

import (
	"testing"
	"time"
)

func TestNew_ZeroSize_ReturnsError(t *testing.T) {
	_, err := New(0)
	if err == nil {
		t.Fatal("expected error for zero size")
	}
}

func TestNew_NegativeSize_ReturnsError(t *testing.T) {
	_, err := New(-time.Second)
	if err == nil {
		t.Fatal("expected error for negative size")
	}
}

func TestNew_ValidSize_NoError(t *testing.T) {
	_, err := New(time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCount_EmptyWindow_ReturnsZero(t *testing.T) {
	w, _ := New(time.Minute)
	if c := w.Count("svc", time.Now()); c != 0 {
		t.Fatalf("expected 0, got %d", c)
	}
}

func TestAdd_CountReflectsEntries(t *testing.T) {
	w, _ := New(time.Minute)
	now := time.Now()
	w.Add("svc", now)
	w.Add("svc", now)
	if c := w.Count("svc", now); c != 2 {
		t.Fatalf("expected 2, got %d", c)
	}
}

func TestCount_EvictsOldEntries(t *testing.T) {
	w, _ := New(time.Minute)
	old := time.Now().Add(-2 * time.Minute)
	w.Add("svc", old)
	w.Add("svc", old)
	now := time.Now()
	w.Add("svc", now)
	if c := w.Count("svc", now); c != 1 {
		t.Fatalf("expected 1 after eviction, got %d", c)
	}
}

func TestServices_ReturnsTrackedNames(t *testing.T) {
	w, _ := New(time.Minute)
	now := time.Now()
	w.Add("alpha", now)
	w.Add("beta", now)
	svcs := w.Services()
	if len(svcs) != 2 {
		t.Fatalf("expected 2 services, got %d", len(svcs))
	}
}

func TestCount_IndependentPerService(t *testing.T) {
	w, _ := New(time.Minute)
	now := time.Now()
	w.Add("a", now)
	w.Add("a", now)
	w.Add("b", now)
	if c := w.Count("a", now); c != 2 {
		t.Fatalf("expected 2 for a, got %d", c)
	}
	if c := w.Count("b", now); c != 1 {
		t.Fatalf("expected 1 for b, got %d", c)
	}
}
