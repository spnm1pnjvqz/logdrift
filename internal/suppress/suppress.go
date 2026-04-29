// Package suppress drops repeated identical log lines within a sliding
// count window. Unlike debounce (which is time-based), suppress counts
// occurrences: once a line has been seen N times it is silently dropped
// until a different line resets the counter for that service.
package suppress

import (
	"errors"
	"sync"

	"github.com/user/logdrift/internal/runner"
)

// Suppressor tracks per-service repetition state.
type Suppressor struct {
	mu      sync.Mutex
	maxReps int
	counts  map[string]int    // service -> consecutive count
	last    map[string]string // service -> last seen text
}

// New returns a Suppressor that allows at most maxReps consecutive identical
// lines per service. maxReps must be >= 1.
func New(maxReps int) (*Suppressor, error) {
	if maxReps < 1 {
		return nil, errors.New("suppress: maxReps must be >= 1")
	}
	return &Suppressor{
		maxReps: maxReps,
		counts:  make(map[string]int),
		last:    make(map[string]string),
	}, nil
}

// Allow returns true when the line should be forwarded downstream.
func (s *Suppressor) Allow(line runner.LogLine) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.last[line.Service] != line.Text {
		s.last[line.Service] = line.Text
		s.counts[line.Service] = 1
		return true
	}
	s.counts[line.Service]++
	return s.counts[line.Service] <= s.maxReps
}

// Apply reads from src, suppresses over-repeated lines, and writes survivors
// to the returned channel. The output channel is closed when src is closed or
// ctx is done.
func Apply(ctx interface{ Done() <-chan struct{} }, s *Suppressor, src <-chan runner.LogLine) <-chan runner.LogLine {
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
				if s.Allow(line) {
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
