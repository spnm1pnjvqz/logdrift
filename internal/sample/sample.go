// Package sample provides periodic sampling of log lines, emitting one
// line every N lines received per service.
package sample

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/yourorg/logdrift/internal/runner"
)

// Sampler emits one out of every N lines per service.
type Sampler struct {
	n       int
	mu      sync.Mutex
	counters map[string]int
}

// New creates a Sampler that passes through every Nth line.
// n must be >= 1; n=1 passes all lines.
func New(n int) (*Sampler, error) {
	if n < 1 {
		return nil, fmt.Errorf("sample: n must be >= 1, got %d", n)
	}
	return &Sampler{
		n:        n,
		counters: make(map[string]int),
	}, nil
}

// ErrZeroN is returned when n is zero.
var ErrZeroN = errors.New("sample: n must be >= 1")

// keep returns true if this line should be forwarded.
func (s *Sampler) keep(service string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[service]++
	return s.counters[service]%s.n == 0
}

// Apply reads lines from src and forwards every Nth line to the returned channel.
// The output channel is closed when src is closed or ctx is cancelled.
func (s *Sampler) Apply(ctx context.Context, src <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-src:
				if !ok {
					return
				}
				if s.keep(line.Service) {
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
