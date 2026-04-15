// Package ratelimit provides a token-bucket rate limiter for log lines,
// allowing users to cap the number of lines processed per second per service.
package ratelimit

import (
	"context"
	"time"

	"github.com/user/logdrift/internal/runner"
)

// Limiter holds configuration for the rate limiter.
type Limiter struct {
	// linesPerSec is the maximum number of lines allowed per second.
	// A value of 0 means unlimited.
	linesPerSec int
	ticker      *time.Ticker
	tokens      chan struct{}
}

// New creates a new Limiter. If linesPerSec is 0, no limiting is applied.
func New(linesPerSec int) (*Limiter, error) {
	if linesPerSec < 0 {
		return nil, fmt.Errorf("ratelimit: linesPerSec must be >= 0, got %d", linesPerSec)
	}
	return &Limiter{linesPerSec: linesPerSec}, nil
}

// Apply wraps an input channel of log lines with rate limiting, returning a
// new channel that emits lines at most linesPerSec times per second.
// If linesPerSec is 0, input is forwarded without delay.
func (l *Limiter) Apply(ctx context.Context, in <-chan runner.Line) <-chan runner.Line {
	out := make(chan runner.Line)

	go func() {
		defer close(out)

		if l.linesPerSec == 0 {
			// Unlimited: pass through directly.
			for {
				select {
				case <-ctx.Done():
					return
				case line, ok := <-in:
					if !ok {
						return
					}
					out <- line
				}
			}
		}

		interval := time.Second / time.Duration(l.linesPerSec)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-in:
				if !ok {
					return
				}
				// Wait for the next tick before forwarding.
				select {
				case <-ticker.C:
				case <-ctx.Done():
					return
				}
				out <- line
			}
		}
	}()

	return out
}
