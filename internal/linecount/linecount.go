// Package linecount tracks per-service line counts over a sliding window.
package linecount

import (
	"errors"
	"sync"

	"github.com/user/logdrift/internal/runner"
)

// Counter tracks how many lines each service has emitted.
type Counter struct {
	mu     sync.Mutex
	counts map[string]int64
}

// New returns an initialised Counter.
func New() *Counter {
	return &Counter{counts: make(map[string]int64)}
}

// Add increments the count for the given service by 1.
func (c *Counter) Add(service string) error {
	if service == "" {
		return errors.New("linecount: service name must not be empty")
	}
	c.mu.Lock()
	c.counts[service]++
	c.mu.Unlock()
	return nil
}

// Get returns the current count for a service.
func (c *Counter) Get(service string) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counts[service]
}

// Snapshot returns a copy of all counts.
func (c *Counter) Snapshot() map[string]int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]int64, len(c.counts))
	for k, v := range c.counts {
		out[k] = v
	}
	return out
}

// Apply consumes lines from ch, increments the counter for each line's
// service, and forwards every line to the returned channel.
func (c *Counter) Apply(ctx interface{ Done() <-chan struct{} }, ch <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-ch:
				if !ok {
					return
				}
				_ = c.Add(line.Service)
				select {
				case out <- line:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
