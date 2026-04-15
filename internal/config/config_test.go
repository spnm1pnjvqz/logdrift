package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	tmp := filepath.Join(t.TempDir(), "logdrift.yaml")
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTemp: %v", err)
	}
	return tmp
}

func TestLoad_Valid(t *testing.T) {
	path := writeTemp(t, `
services:
  - name: api
    command: "tail -f /var/log/api.log"
    color: cyan
  - name: worker
    command: "journalctl -fu worker"
diff_mode: word
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(cfg.Services))
	}
	if cfg.DiffMode != "word" {
		t.Errorf("expected diff_mode=word, got %q", cfg.DiffMode)
	}
}

func TestLoad_DefaultDiffMode(t *testing.T) {
	path := writeTemp(t, `
services:
  - name: svc
    command: "echo hello"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DiffMode != "line" {
		t.Errorf("expected default diff_mode=line, got %q", cfg.DiffMode)
	}
}

func TestLoad_MissingCommand(t *testing.T) {
	path := writeTemp(t, `
services:
  - name: broken
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing command, got nil")
	}
}

func TestLoad_DuplicateName(t *testing.T) {
	path := writeTemp(t, `
services:
  - name: dup
    command: "echo a"
  - name: dup
    command: "echo b"
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate service name, got nil")
	}
}

func TestLoad_InvalidDiffMode(t *testing.T) {
	path := writeTemp(t, `
services:
  - name: svc
    command: "echo hi"
diff_mode: char
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid diff_mode, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/logdrift.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
