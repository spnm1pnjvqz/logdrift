// Package throttle provides a token-bucket throttle that limits how many
// log lines per second are forwarded downstream.
package throttle

import (
	"context"
	"errors"
	"time"

	"github.com/user/logdrift/internal/runner"
)

// Throttle holds the configuration for rate-limiting a log line channel.
type Throttle struct {
	interval time.Duration
}

// New returns a Throttle that allows at most linesPerSec lines per second.
// linesPerSec must be >= 1.
func New(linesPerSec int) (*Throttle, error) {
	if linesPerSec < 1 {
		return nil, errors.New("throttle: linesPerSec must be >= 1")
	}
	return &Throttle{interval: time.Second / time.Duration(linesPerSec)}, nil
}

// Apply reads lines from in and forwards them to the returned channel,
// inserting a minimum delay of t.interval between each forwarded line.
// The output channel is closed when ctx is cancelled or in is closed.
func (t *Throttle) Apply(ctx context.Context, in <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		ticker := time.NewTicker(t.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-in:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
				}
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
