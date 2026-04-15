// Package truncate provides utilities for truncating long log lines
// before they are passed downstream in the pipeline.
package truncate

import (
	"context"
	"fmt"

	"github.com/user/logdrift/internal/runner"
)

const defaultMaxBytes = 1024

// Truncator holds configuration for line truncation.
type Truncator struct {
	maxBytes int
	suffix   string
}

// New creates a Truncator that clips lines longer than maxBytes.
// If maxBytes is 0 the default of 1024 is used.
// suffix is appended to truncated lines (e.g. "...").
func New(maxBytes int, suffix string) (*Truncator, error) {
	if maxBytes < 0 {
		return nil, fmt.Errorf("truncate: maxBytes must be >= 0, got %d", maxBytes)
	}
	if maxBytes == 0 {
		maxBytes = defaultMaxBytes
	}
	if suffix == "" {
		suffix = "..."
	}
	return &Truncator{maxBytes: maxBytes, suffix: suffix}, nil
}

// Truncate clips a single line to maxBytes (in bytes).
// If the line is within the limit it is returned unchanged.
func (t *Truncator) Truncate(line string) string {
	if len(line) <= t.maxBytes {
		return line
	}
	// Clip at maxBytes and append suffix.
	clipped := line[:t.maxBytes]
	return clipped + t.suffix
}

// Apply reads LogLines from in, truncates each message, and writes the
// (possibly modified) line to the returned channel. The output channel
// is closed when ctx is cancelled or in is closed.
func (t *Truncator) Apply(ctx context.Context, in <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case ll, ok := <-in:
				if !ok {
					return
				}
				ll.Line = t.Truncate(ll.Line)
				select {
				case out <- ll:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
