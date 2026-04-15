package tail_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/logdrift/internal/tail"
)

func writeLine(t *testing.T, f *os.File, line string) {
	t.Helper()
	_, err := f.WriteString(line + "\n")
	if err != nil {
		t.Fatalf("writeLine: %v", err)
	}
}

func TestNew_MissingFile_ReturnsError(t *testing.T) {
	_, err := tail.New("/nonexistent/path/to/file.log")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestNew_ExistingFile_NoError(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test.log")
	os.WriteFile(tmp, []byte("existing\n"), 0644)
	_, err := tail.New(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTail_EmitsNewLines(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test.log")
	f, err := os.Create(tmp)
	if err != nil {
		t.Fatal(err)
	}

	tr, err := tail.New(tmp)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := tr.Tail(ctx, "svc")
	if err != nil {
		t.Fatal(err)
	}

	// Give goroutine time to seek.
	time.Sleep(50 * time.Millisecond)
	writeLine(t, f, "hello world")
	writeLine(t, f, "second line")
	f.Close()

	var got []string
	timeout := time.After(2 * time.Second)
	for len(got) < 2 {
		select {
		case line, ok := <-ch:
			if !ok {
				t.Fatal("channel closed before receiving expected lines")
			}
			got = append(got, line.Text)
		case <-timeout:
			t.Fatalf("timeout waiting for lines, got %d", len(got))
		}
	}

	if got[0] != "hello world\n" {
		t.Errorf("line 0: got %q, want %q", got[0], "hello world\n")
	}
}

func TestTail_CancelClosesChannel(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test.log")
	os.WriteFile(tmp, []byte{}, 0644)

	tr, _ := tail.New(tmp)
	ctx, cancel := context.WithCancel(context.Background())
	ch, _ := tr.Tail(ctx, "svc")

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-ch:
		// closed or received, both acceptable
	case <-time.After(time.Second):
		t.Fatal("channel not closed after cancel")
	}
}
