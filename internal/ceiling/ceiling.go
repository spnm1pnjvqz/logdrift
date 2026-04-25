// Package ceiling provides a line-count ceiling per service: once a service
// has emitted N lines the stream for that service is silently dropped.
package ceiling

import (
	"errors"
	"sync"

	"github.com/user/logdrift/internal/runner"
)

// Ceiling drops lines from a service once its per-service count reaches Max.
type Ceiling struct {
	max    int
	mu     sync.Mutex
	counts map[string]int
}

// New creates a Ceiling that allows at most max lines per service.
// max must be >= 1.
func New(max int) (*Ceiling, error) {
	if max < 1 {
		return nil, errors.New("ceiling: max must be >= 1")
	}
	return &Ceiling{
		max:    max,
		counts: make(map[string]int),
	}, nil
}

// Allow returns true when the line should be forwarded (count not yet reached).
func (c *Ceiling) Allow(line runner.LogLine) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[line.Service]++
	return c.counts[line.Service] <= c.max
}

// Apply reads from src, forwards lines that are still under the ceiling, and
// closes the returned channel when src is closed or ctx is done.
func Apply(c *Ceiling, src <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		for line := range src {
			if c.Allow(line) {
				out <- line
			}
		}
	}()
	return out
}
