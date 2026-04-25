// Package offset stamps each log line with a byte offset counter,
// tracking how many bytes have been seen from a given service.
package offset

import (
	"fmt"
	"sync"

	"github.com/user/logdrift/internal/runner"
)

// Stamper tracks cumulative byte offsets per service.
type Stamper struct {
	mu      sync.Mutex
	offsets map[string]int64
	format  string
}

// New returns a Stamper. format is a fmt-style template that must contain
// exactly one %d verb (the offset) and one %s verb (the original text),
// e.g. "[%010d] %s". If format is empty a default is used.
func New(format string) (*Stamper, error) {
	if format == "" {
		format = "[%010d] %s"
	}
	return &Stamper{
		offsets: make(map[string]int64),
		format:  format,
	}, nil
}

// Stamp prepends the cumulative byte offset to line.Text and returns the
// updated line. The offset is incremented by len(line.Text)+1 (for newline).
func (s *Stamper) Stamp(line runner.LogLine) runner.LogLine {
	s.mu.Lock()
	current := s.offsets[line.Service]
	s.offsets[line.Service] = current + int64(len(line.Text)) + 1
	s.mu.Unlock()

	line.Text = fmt.Sprintf(s.format, current, line.Text)
	return line
}

// Apply reads lines from src, stamps each one, and sends the result to the
// returned channel. The output channel is closed when src is closed or ctx
// is cancelled.
func (s *Stamper) Apply(ctx interface{ Done() <-chan struct{} }, src <-chan runner.LogLine) <-chan runner.LogLine {
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

// Current returns the current byte offset for the given service.
func (s *Stamper) Current(service string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.offsets[service]
}

// Reset zeroes the byte offset for the given service.
func (s *Stamper) Reset(service string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.offsets[service] = 0
}
