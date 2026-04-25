// Package batch groups log lines into fixed-size or time-bounded batches,
// emitting a slice of lines once the batch is full or the flush interval elapses.
package batch

import (
	"errors"
	"time"

	"github.com/user/logdrift/internal/runner"
)

// Batch holds configuration for the batcher.
type Batch struct {
	size     int
	interval time.Duration
}

// New creates a Batch with the given size and flush interval.
// size must be >= 1; interval must be > 0.
func New(size int, interval time.Duration) (*Batch, error) {
	if size < 1 {
		return nil, errors.New("batch: size must be at least 1")
	}
	if interval <= 0 {
		return nil, errors.New("batch: interval must be positive")
	}
	return &Batch{size: size, interval: interval}, nil
}

// Apply reads lines from src and emits []runner.LogLine batches on the returned
// channel. A batch is flushed when it reaches b.size lines or b.interval
// elapses since the first line in the current batch, whichever comes first.
// The output channel is closed when ctx is done or src is closed.
func (b *Batch) Apply(ctx interface{ Done() <-chan struct{} }, src <-chan runner.LogLine) <-chan []runner.LogLine {
	out := make(chan []runner.LogLine)
	go func() {
		defer close(out)
		buf := make([]runner.LogLine, 0, b.size)
		timer := time.NewTimer(b.interval)
		timer.Stop()
		flush := func() {
			if len(buf) == 0 {
				return
			}
			copy := append([]runner.LogLine(nil), buf...)
			select {
			case out <- copy:
			case <-ctx.Done():
				return
			}
			buf = buf[:0]
		}
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-src:
				if !ok {
					timer.Stop()
					flush()
					return
				}
				if len(buf) == 0 {
					timer.Reset(b.interval)
				}
				buf = append(buf, line)
				if len(buf) >= b.size {
					timer.Stop()
					flush()
				}
			case <-timer.C:
				flush()
			}
		}
	}()
	return out
}
