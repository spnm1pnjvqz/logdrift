// Package head limits a log line channel to the first N lines per service.
package head

import (
	"context"
	"errors"
	"sync"

	"github.com/user/logdrift/internal/runner"
)

// Limiter tracks per-service line counts and drops lines once the limit is reached.
type Limiter struct {
	max    int
	mu     sync.Mutex
	counts map[string]int
}

// New creates a Limiter that passes at most n lines per service.
// Returns an error if n is less than 1.
func New(n int) (*Limiter, error) {
	if n < 1 {
		return nil, errors.New("head: n must be >= 1")
	}
	return &Limiter{max: n, counts: make(map[string]int)}, nil
}

// Allow reports whether the line should be forwarded.
func (l *Limiter) Allow(line runner.LogLine) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.counts[line.Service]++
	return l.counts[line.Service] <= l.max
}

// Apply reads from in and forwards only the first n lines per service to the
// returned channel. The output channel is closed when in is closed or ctx is
// cancelled.
func (l *Limiter) Apply(ctx context.Context, in <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-in:
				if !ok {
					return
				}
				if l.Allow(line) {
					select {
					case out <- line:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out
}
