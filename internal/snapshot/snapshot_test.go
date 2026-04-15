package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/snapshot"
)

func TestNew_EmptyEntries(t *testing.T) {
	s := snapshot.New()
	if s == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if len(s.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(s.Entries))
	}
	if s.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
}

func TestAdd_AppendsEntries(t *testing.T) {
	s := snapshot.New()
	s.Add("svc-a", "hello world")
	s.Add("svc-b", "another line")

	if len(s.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(s.Entries))
	}
	if s.Entries[0].Service != "svc-a" {
		t.Errorf("expected service svc-a, got %s", s.Entries[0].Service)
	}
	if s.Entries[1].Line != "another line" {
		t.Errorf("unexpected line: %s", s.Entries[1].Line)
	}
}

func TestAdd_CapturedAtSet(t *testing.T) {
	before := time.Now()
	s := snapshot.New()
	s.Add("svc", "line")
	after := time.Now()

	ca := s.Entries[0].CapturedAt
	if ca.Before(before) || ca.After(after) {
		t.Errorf("CapturedAt %v out of range [%v, %v]", ca, before, after)
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	s := snapshot.New()
	s.Add("alpha", "line one")
	s.Add("beta", "line two")

	dir := t.TempDir()
	path := filepath.Join(dir, "snap.json")

	if err := s.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := snapshot.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(loaded.Entries) != 2 {
		t.Fatalf("expected 2 entries after load, got %d", len(loaded.Entries))
	}
	if loaded.Entries[0].Service != "alpha" {
		t.Errorf("unexpected service: %s", loaded.Entries[0].Service)
	}
	if loaded.Entries[1].Line != "line two" {
		t.Errorf("unexpected line: %s", loaded.Entries[1].Line)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := snapshot.Load("/nonexistent/path/snap.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestSave_BadPath(t *testing.T) {
	s := snapshot.New()
	err := s.Save("/nonexistent/dir/snap.json")
	if err == nil {
		t.Fatal("expected error for bad path")
	}
}

func TestLoad_CorruptJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("not json{"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := snapshot.Load(path)
	if err == nil {
		t.Fatal("expected error for corrupt JSON")
	}
}
