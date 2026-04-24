// Package overflow provides a channel-based back-pressure limiter that drops
// or blocks lines when a downstream consumer is too slow.
package overflow

import (
	"errors"
	"fmt"

	"github.com/user/logdrift/internal/runner"
)

// Policy controls what happens when the output buffer is full.
type Policy int

const (
	// Drop silently discards lines that cannot be queued.
	Drop Policy = iota
	// Block causes the Apply goroutine to wait until space is available.
	Block
)

const defaultCapacity = 256

// Limiter holds configuration for the overflow handler.
type Limiter struct {
	capacity int
	policy   Policy
}

// New creates a Limiter with the given buffer capacity and overflow policy.
// capacity <= 0 uses the default of 256.
func New(capacity int, policy Policy) (*Limiter, error) {
	if policy != Drop && policy != Block {
		return nil, errors.New("overflow: unknown policy")
	}
	if capacity < 0 {
		return nil, fmt.Errorf("overflow: capacity must be >= 0, got %d", capacity)
	}
	if capacity == 0 {
		capacity = defaultCapacity
	}
	return &Limiter{capacity: capacity, policy: policy}, nil
}

// Apply wraps src in a buffered channel. Lines are forwarded according to the
// configured policy when the buffer is full. The returned channel is closed
// when ctx is done or src is closed.
func (l *Limiter) Apply(ctx interface{ Done() <-chan struct{} }, src <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine, l.capacity)
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
				if l.policy == Block {
					select {
					case out <- line:
					case <-ctx.Done():
						return
					}
				} else {
					// Drop policy: non-blocking send
					select {
					case out <- line:
					default:
					}
				}
			}
		}
	}()
	return out
}
