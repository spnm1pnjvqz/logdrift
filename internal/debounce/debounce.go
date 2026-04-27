// Package debounce suppresses rapid bursts of identical log lines from the
// same service, emitting only the first occurrence within a configurable
// quiet window. Subsequent identical lines restart the timer.
package debounce

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/user/logdrift/internal/runner"
)

// Debouncer holds per-service timers and tracks the last seen text.
type Debouncer struct {
	window time.Duration
	mu     sync.Mutex
	state  map[string]*entry
}

type entry struct {
	lastText string
	timer    *time.Timer
	suppress bool
}

// New creates a Debouncer with the given quiet window.
// Returns an error if window is <= 0.
func New(window time.Duration) (*Debouncer, error) {
	if window <= 0 {
		return nil, fmt.Errorf("debounce: window must be positive, got %s", window)
	}
	return &Debouncer{
		window: window,
		state:  make(map[string]*entry),
	}, nil
}

// Allow returns true if the line should be forwarded downstream.
// The first occurrence of a text/service pair is always forwarded.
// Repeated identical lines within the quiet window are suppressed.
func (d *Debouncer) Allow(line runner.LogLine) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	e, ok := d.state[line.Service]
	if !ok || e.lastText != line.Text {
		// New service or different text — always forward and arm timer.
		if ok && e.timer != nil {
			e.timer.Stop()
		}
		timer := time.AfterFunc(d.window, func() {
			d.mu.Lock()
			if s, exists := d.state[line.Service]; exists {
				s.suppress = false
			}
			d.mu.Unlock()
		})
		d.state[line.Service] = &entry{lastText: line.Text, timer: timer, suppress: true}
		return true
	}
	// Same text within window — suppress and reset timer.
	if e.suppress {
		e.timer.Reset(d.window)
		return false
	}
	// Window expired for same text — forward once more and re-arm.
	e.suppress = true
	e.timer.Reset(d.window)
	return true
}

// Apply reads lines from src, forwarding only those that pass Allow.
// The output channel is closed when src is closed or ctx is cancelled.
func Apply(ctx context.Context, d *Debouncer, src <-chan runner.LogLine) <-chan runner.LogLine {
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
				if d.Allow(line) {
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
