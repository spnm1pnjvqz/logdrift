// Package window provides a sliding time-window counter for log lines per service.
package window

import (
	"errors"
	"sync"
	"time"
)

// Entry holds a timestamp for a single observed log line.
type Entry struct {
	At time.Time
}

// Window tracks log-line counts within a rolling duration.
type Window struct {
	mu       sync.Mutex
	size     time.Duration
	buckets  map[string][]Entry
}

// New creates a Window with the given rolling duration.
func New(size time.Duration) (*Window, error) {
	if size <= 0 {
		return nil, errors.New("window: size must be positive")
	}
	return &Window{
		size:    size,
		buckets: make(map[string][]Entry),
	}, nil
}

// Add records a new log line for service at time t.
func (w *Window) Add(service string, t time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buckets[service] = append(w.buckets[service], Entry{At: t})
	w.evict(service, t)
}

// Count returns the number of lines for service within the window ending at now.
func (w *Window) Count(service string, now time.Time) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.evict(service, now)
	return len(w.buckets[service])
}

// Services returns all service names currently tracked.
func (w *Window) Services() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make([]string, 0, len(w.buckets))
	for k := range w.buckets {
		out = append(out, k)
	}
	return out
}

// evict removes entries older than w.size relative to now. Caller must hold mu.
func (w *Window) evict(service string, now time.Time) {
	cutoff := now.Add(-w.size)
	entries := w.buckets[service]
	i := 0
	for i < len(entries) && entries[i].At.Before(cutoff) {
		i++
	}
	w.buckets[service] = entries[i:]
}
