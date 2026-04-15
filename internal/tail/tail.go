// Package tail provides utilities for tailing files and emitting
// log lines into a channel, similar to `tail -f`.
package tail

import (
	"bufio"
	"context"
	"io"
	"os"
	"time"

	"github.com/user/logdrift/internal/runner"
)

const pollInterval = 200 * time.Millisecond

// Tailer tails a file and emits lines.
type Tailer struct {
	path string
}

// New creates a new Tailer for the given file path.
func New(path string) (*Tailer, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	return &Tailer{path: path}, nil
}

// Tail opens the file, seeks to the end, and emits new lines as they
// are appended. The returned channel is closed when ctx is cancelled
// or an unrecoverable read error occurs.
func (t *Tailer) Tail(ctx context.Context, service string) (<-chan runner.LogLine, error) {
	f, err := os.Open(t.path)
	if err != nil {
		return nil, err
	}

	// Seek to end so we only emit new content.
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		f.Close()
		return nil, err
	}

	ch := make(chan runner.LogLine, 64)
	go func() {
		defer close(ch)
		defer f.Close()
		reader := bufio.NewReader(f)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			line, err := reader.ReadString('\n')
			if len(line) > 0 {
				ch <- runner.LogLine{Service: service, Text: line}
			}
			if err != nil {
				if err == io.EOF {
					select {
					case <-ctx.Done():
						return
					case <-time.After(pollInterval):
					}
					continue
				}
				return
			}
		}
	}()
	return ch, nil
}
