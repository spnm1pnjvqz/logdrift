// Package join provides a transformer that joins consecutive log lines
// from the same service using a configurable separator.
package join

import (
	"context"
	"fmt"
	"strings"

	"github.com/yourorg/logdrift/internal/runner"
)

// Joiner accumulates consecutive lines from the same service and emits
// a single joined line once a flush condition is met (max count or separator match).
type Joiner struct {
	sep      string
	maxLines int
}

// New creates a Joiner that joins up to maxLines consecutive lines from the
// same service using sep as the in-between separator.
// maxLines must be >= 2; sep must be non-empty.
func New(sep string, maxLines int) (*Joiner, error) {
	if sep == "" {
		return nil, fmt.Errorf("join: separator must not be empty")
	}
	if maxLines < 2 {
		return nil, fmt.Errorf("join: maxLines must be >= 2, got %d", maxLines)
	}
	return &Joiner{sep: sep, maxLines: maxLines}, nil
}

// Apply reads lines from in, buffers consecutive lines per service, and emits
// a joined line whenever the service changes or maxLines is reached.
func (j *Joiner) Apply(ctx context.Context, in <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		var (
			curService string
			buf        []string
			last       runner.LogLine
		)
		flush := func() {
			if len(buf) == 0 {
				return
			}
			last.Text = strings.Join(buf, j.sep)
			select {
			case out <- last:
			case <-ctx.Done():
			}
			buf = buf[:0]
		}
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-in:
				if !ok {
					flush()
					return
				}
				if line.Service != curService && curService != "" {
					flush()
				}
				curService = line.Service
				last = line
				buf = append(buf, line.Text)
				if len(buf) >= j.maxLines {
					flush()
				}
			}
		}
	}()
	return out
}
