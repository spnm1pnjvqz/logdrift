// Package aggregate groups log lines by a key (e.g. service name or regex
// capture) and emits periodic summaries over a sliding window.
package aggregate

import (
	"sync"
	"time"

	"github.com/user/logdrift/internal/runner"
)

// Summary holds the count of lines seen for a given key within a window.
type Summary struct {
	Key       string
	Count     int
	WindowEnd time.Time
}

// Aggregator counts log lines per service within a rolling window.
type Aggregator struct {
	mu       sync.Mutex
	counts   map[string]int
	window   time.Duration
	ticker   *time.Ticker
	output   chan Summary
}

// New creates an Aggregator that emits summaries every window duration.
// window must be positive; otherwise ErrInvalidWindow is returned.
func New(window time.Duration) (*Aggregator, error) {
	if window <= 0 {
		return nil, ErrInvalidWindow
	}
	return &Aggregator{
		counts: make(map[string]int),
		window: window,
		output: make(chan Summary, 64),
	}, nil
}

// Apply consumes lines from in, tallies them by service, and flushes
// summaries on each tick. The output channel is closed when ctx is done.
func (a *Aggregator) Apply(in <-chan runner.LogLine) <-chan Summary {
	a.ticker = time.NewTicker(a.window)
	go func() {
		defer close(a.output)
		defer a.ticker.Stop()
		for {
			select {
			case line, ok := <-in:
				if !ok {
					a.flush()
					return
				}
				a.mu.Lock()
				a.counts[line.Service]++
				a.mu.Unlock()
			case t := <-a.ticker.C:
				a.flushAt(t)
			}
		}
	}()
	return a.output
}

func (a *Aggregator) flush() {
	a.flushAt(time.Now())
}

func (a *Aggregator) flushAt(t time.Time) {
	a.mu.Lock()
	snap := a.counts
	a.counts = make(map[string]int)
	a.mu.Unlock()
	for k, v := range snap {
		a.output <- Summary{Key: k, Count: v, WindowEnd: t}
	}
}
