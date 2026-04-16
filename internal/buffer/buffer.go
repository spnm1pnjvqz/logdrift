// Package buffer provides a fixed-size ring buffer for log lines,
// retaining the N most recent lines per service.
package buffer

import (
	"fmt"
	"sync"

	"github.com/user/logdrift/internal/runner"
)

const defaultCapacity = 100

// Buffer holds the last N log lines per service.
type Buffer struct {
	mu       sync.RWMutex
	cap      int
	lines    map[string][]runner.LogLine
}

// New creates a Buffer that retains at most capacity lines per service.
// If capacity is <= 0, defaultCapacity is used.
func New(capacity int) (*Buffer, error) {
	if capacity < 0 {
		return nil, fmt.Errorf("buffer: capacity must be >= 0, got %d", capacity)
	}
	if capacity == 0 {
		capacity = defaultCapacity
	}
	return &Buffer{
		cap:   capacity,
		lines: make(map[string][]runner.LogLine),
	}, nil
}

// Add appends a log line to the ring buffer for its service.
// When the buffer is full the oldest line is evicted.
func (b *Buffer) Add(line runner.LogLine) {
	b.mu.Lock()
	defer b.mu.Unlock()

	svc := line.Service
	buf := b.lines[svc]
	if len(buf) >= b.cap {
		buf = buf[1:]
	}
	b.lines[svc] = append(buf, line)
}

// Get returns a copy of the buffered lines for the given service.
// Returns nil if the service has no recorded lines.
func (b *Buffer) Get(service string) []runner.LogLine {
	b.mu.RLock()
	defer b.mu.RUnlock()

	src := b.lines[service]
	if len(src) == 0 {
		return nil
	}
	out := make([]runner.LogLine, len(src))
	copy(out, src)
	return out
}

// Services returns the names of all services that have at least one line.
func (b *Buffer) Services() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	svcs := make([]string, 0, len(b.lines))
	for svc := range b.lines {
		svcs = append(svcs, svc)
	}
	return svcs
}

// Len returns the number of buffered lines for the given service.
func (b *Buffer) Len(service string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.lines[service])
}
