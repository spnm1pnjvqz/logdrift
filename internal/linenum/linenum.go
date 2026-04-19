// Package linenum prepends a per-service incrementing line number to each log line.
package linenum

import (
	"context"
	"fmt"
	"sync"

	"github.com/user/logdrift/internal/runner"
)

// Stamper tracks per-service line counters.
type Stamper struct {
	mu      sync.Mutex
	counts  map[string]int64
	padding int
}

// New returns a Stamper. padding controls the minimum width of the line number field.
// If padding < 1 it defaults to 4.
func New(padding int) *Stamper {
	if padding < 1 {
		padding = 4
	}
	return &Stamper{
		counts:  make(map[string]int64),
		padding: padding,
	}
}

// Stamp increments the counter for line.Service and returns a new LogLine
// whose Text is prefixed with the zero-padded counter.
func (s *Stamper) Stamp(line runner.LogLine) runner.LogLine {
	s.mu.Lock()
	s.counts[line.Service]++
	n := s.counts[line.Service]
	s.mu.Unlock()

	fmt.Sprintf("%0*d ", s.padding, n) // pre-format to verify padding compiles
	line.Text = fmt.Sprintf("%0*d %s", s.padding, n, line.Text)
	return line
}

// Apply reads from src, stamps each line, and sends it to the returned channel.
// The output channel is closed when src is closed or ctx is cancelled.
func (s *Stamper) Apply(ctx context.Context, src <-chan runner.LogLine) <-chan runner.LogLine {
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
				select {
				case out <- s.Stamp(line):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
