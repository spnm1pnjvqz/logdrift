// Package retry provides a channel-based retry wrapper that re-emits
// log lines after a configurable back-off when a source channel closes
// unexpectedly before the context is done.
package retry

import (
	"context"
	"errors"
	"time"

	"github.com/user/logdrift/internal/runner"
)

// ErrInvalidMaxAttempts is returned when MaxAttempts is less than 1.
var ErrInvalidMaxAttempts = errors.New("retry: MaxAttempts must be >= 1")

// Config holds retry policy settings.
type Config struct {
	MaxAttempts int
	Delay       time.Duration
}

// Retryer wraps a line source factory and retries on unexpected close.
type Retryer struct {
	cfg Config
}

// New validates cfg and returns a Retryer.
func New(cfg Config) (*Retryer, error) {
	if cfg.MaxAttempts < 1 {
		return nil, ErrInvalidMaxAttempts
	}
	if cfg.Delay < 0 {
		cfg.Delay = 0
	}
	return &Retryer{cfg: cfg}, nil
}

// Apply wraps a factory function that returns a line channel. It retries up to
// MaxAttempts times when the channel closes while ctx is still active.
func (r *Retryer) Apply(ctx context.Context, factory func(ctx context.Context) (<-chan runner.LogLine, error)) (<-chan runner.LogLine, error) {
	out := make(chan runner.LogLine)
	src, err := factory(ctx)
	if err != nil {
		return nil, err
	}
	go func() {
		defer close(out)
		attempts := 0
		ch := src
		for {
			for line := range ch {
				select {
				case out <- line:
				case <-ctx.Done():
					return
				}
			}
			// channel closed
			select {
			case <-ctx.Done():
				return
			default:
			}
			attempts++
			if attempts >= r.cfg.MaxAttempts {
				return
			}
			select {
			case <-time.After(r.cfg.Delay):
			case <-ctx.Done():
				return
			}
			ch, err = factory(ctx)
			if err != nil {
				return
			}
		}
	}()
	return out, nil
}
