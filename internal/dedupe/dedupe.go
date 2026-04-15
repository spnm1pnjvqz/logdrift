// Package dedupe provides a pipeline stage that suppresses consecutive
// duplicate log lines from the same service.
package dedupe

import (
	"context"
	"sync"

	"github.com/user/logdrift/internal/runner"
)

// Deduper holds the last-seen line per service label.
type Deduper struct {
	mu   sync.Mutex
	last map[string]string
}

// New returns a new Deduper with an empty history.
func New() *Deduper {
	return &Deduper{last: make(map[string]string)}
}

// IsDuplicate reports whether line is identical to the previous line seen
// for the given service label. If it is not a duplicate the line is recorded.
func (d *Deduper) IsDuplicate(label, line string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if prev, ok := d.last[label]; ok && prev == line {
		return true
	}
	d.last[label] = line
	return false
}

// Reset clears the recorded history for all labels.
func (d *Deduper) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.last = make(map[string]string)
}

// Apply reads LogLines from in, drops consecutive duplicates per service
// label, and forwards the rest to the returned channel. The output channel
// is closed when ctx is cancelled or in is closed.
func Apply(ctx context.Context, d *Deduper, in <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case ll, ok := <-in:
				if !ok {
					return
				}
				if d.IsDuplicate(ll.Service, ll.Line) {
					continue
				}
				select {
				case out <- ll:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
