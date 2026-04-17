// Package coalesce merges bursts of log lines from the same service within a
// short time window into a single aggregated line.
package coalesce

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/user/logdrift/internal/runner"
)

// Coalescer buffers lines per service and flushes them after a quiet period.
type Coalescer struct {
	window time.Duration
}

// New returns a Coalescer that flushes buffered lines after window of silence.
func New(window time.Duration) (*Coalescer, error) {
	if window <= 0 {
		return nil, fmt.Errorf("coalesce: window must be positive, got %v", window)
	}
	return &Coalescer{window: window}, nil
}

// Apply reads lines from in, coalesces bursts per service, and emits merged lines.
func (c *Coalescer) Apply(ctx context.Context, in <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		buf := map[string][]string{}
		timers := map[string]*time.Timer{}
		flush := make(chan string, 32)

		send := func(svc string) {
			lines := buf[svc]
			if len(lines) == 0 {
				return
			}
			merged := strings.Join(lines, " | ")
			delete(buf, svc)
			select {
			case out <- runner.LogLine{Service: svc, Text: merged}:
			case <-ctx.Done():
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case svc := <-flush:
				send(svc)
			case line, ok := <-in:
				if !ok {
					for svc := range buf {
						send(svc)
					}
					return
				}
				buf[line.Service] = append(buf[line.Service], line.Text)
				if t, ok := timers[line.Service]; ok {
					t.Reset(c.window)
				} else {
					svc := line.Service
					timers[svc] = time.AfterFunc(c.window, func() {
						flush <- svc
					})
				}
			}
		}
	}()
	return out
}
