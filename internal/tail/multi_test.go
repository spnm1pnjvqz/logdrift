package tail_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/logdrift/internal/tail"
)

func TestTailAll_NoSources_ReturnsError(t *testing.T) {
	_, err := tail.TailAll(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for empty sources")
	}
}

func TestTailAll_MissingFile_ReturnsError(t *testing.T) {
	sources := []tail.FileSource{
		{Service: "svc", Path: "/no/such/file.log"},
	}
	_, err := tail.TailAll(context.Background(), sources)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestTailAll_MergesMultipleSources(t *testing.T) {
	dir := t.TempDir()

	pathA := filepath.Join(dir, "a.log")
	pathB := filepath.Join(dir, "b.log")
	os.WriteFile(pathA, []byte{}, 0644)
	os.WriteFile(pathB, []byte{}, 0644)

	sources := []tail.FileSource{
		{Service: "alpha", Path: pathA},
		{Service: "beta", Path: pathB},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := tail.TailAll(ctx, sources)
	if err != nil {
		t.Fatalf("TailAll: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	fA, _ := os.OpenFile(pathA, os.O_APPEND|os.O_WRONLY, 0644)
	fB, _ := os.OpenFile(pathB, os.O_APPEND|os.O_WRONLY, 0644)
	fA.WriteString("from alpha\n")
	fB.WriteString("from beta\n")
	fA.Close()
	fB.Close()

	seen := map[string]bool{}
	timeout := time.After(2 * time.Second)
	for len(seen) < 2 {
		select {
		case line, ok := <-ch:
			if !ok {
				t.Fatal("channel closed early")
			}
			seen[line.Service] = true
		case <-timeout:
			t.Fatalf("timeout: only received from services: %v", seen)
		}
	}

	if !seen["alpha"] || !seen["beta"] {
		t.Errorf("did not receive from all services: %v", seen)
	}
}
