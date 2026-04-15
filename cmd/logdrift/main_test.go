package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "logdrift.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTempConfig: %v", err)
	}
	return p
}

func TestRun_MissingConfig(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"logdrift", "/nonexistent/path/logdrift.yaml"}

	if err := run(); err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}

func TestRun_InvalidConfig(t *testing.T) {
	p := writeTempConfig(t, `services: []\ndiff_mode: invalid\n`)

	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"logdrift", p}

	if err := run(); err == nil {
		t.Fatal("expected error for invalid config, got nil")
	}
}

func TestRun_DefaultConfigPath(t *testing.T) {
	// Ensure run() returns an error (missing default file) rather than panicking
	// when no args are provided and logdrift.yaml does not exist.
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"logdrift"}

	// Change to a temp dir so logdrift.yaml definitely doesn't exist.
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(t.TempDir())   //nolint:errcheck

	if err := run(); err == nil {
		t.Fatal("expected error when default config missing, got nil")
	}
}
