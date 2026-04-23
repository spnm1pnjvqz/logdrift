// Package reorder buffers log lines within a short window and emits them
// sorted by their embedded timestamp so that slightly out-of-order lines
// from different services arrive at the display layer in chronological order.
package reorder

import (
	"sort"
	"time"

	"github.com/user/logdrift/internal/runner"
)

// Reorderer holds configuration for the reorder stage.
type Reorderer struct {
	window  time.Duration
	extract func(text string) (time.Time, bool)
}

// New returns a Reorderer that buffers lines for the given window duration
// before flushing them in timestamp order. extract is called on each line's
// Text field; if it returns false the line's arrival time is used instead.
func New(window time.Duration, extract func(string) (time.Time, bool)) (*Reorderer, error) {
	if window <= 0 {
		return nil, errInvalidWindow
	}
	if extract == nil {
		return nil, errNilExtract
	}
	return &Reorderer{window: window, extract: extract}, nil
}

// Apply reads from src, buffers lines for the configured window, and emits
// them in ascending timestamp order. The output channel is closed when src is
// drained or ctx is cancelled.
func (r *Reorderer) Apply(ctx interface{ Done() <-chan struct{} }, src <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine, 64)
	go func() {
		defer close(out)
		timer := time.NewTimer(r.window)
		defer timer.Stop()

		type stamped struct {
			t  time.Time
			ll runner.LogLine
		}
		var buf []stamped

		flush := func() {
			sort.Slice(buf, func(i, j int) bool {
				return buf[i].t.Before(buf[j].t)
			})
			for _, s := range buf {
				select {
				case out <- s.ll:
				case <-ctx.Done():
					return
				}
			}
			buf = buf[:0]
		}

		for {
			select {
			case <-ctx.Done():
				return
			case ll, ok := <-src:
				if !ok {
					flush()
					return
				}
				t, ok2 := r.extract(ll.Text)
				if !ok2 {
					t = time.Now()
				}
				buf = append(buf, stamped{t: t, ll: ll})
			case <-timer.C:
				flush()
				timer.Reset(r.window)
			}
		}
	}()
	return out
}
