// Package since filters log lines to only emit those whose parsed timestamp
// is at or after a given cutoff time.
package since

import (
	"fmt"
	"time"

	"github.com/user/logdrift/internal/runner"
)

// Filter drops lines whose timestamp falls before Cutoff.
type Filter struct {
	Cutoff  time.Time
	Formats []string
}

// New returns a Filter that passes only lines at or after cutoff.
// formats is an ordered list of time.Parse layouts to try.
func New(cutoff time.Time, formats []string) (*Filter, error) {
	if cutoff.IsZero() {
		return nil, fmt.Errorf("since: cutoff time must not be zero")
	}
	if len(formats) == 0 {
		formats = []string{time.RFC3339, "2006-01-02 15:04:05", "Jan 2 15:04:05"}
	}
	return &Filter{Cutoff: cutoff, Formats: formats}, nil
}

// Allow returns true when the line's timestamp is at or after the cutoff.
// Lines whose timestamp cannot be parsed are always allowed through.
func (f *Filter) Allow(line runner.LogLine) bool {
	for _, layout := range f.Formats {
		if t, err := time.Parse(layout, line.Text); err == nil {
			return !t.Before(f.Cutoff)
		}
		// try extracting a prefix of the same length as the layout
		if len(line.Text) >= len(layout) {
			if t, err := time.Parse(layout, line.Text[:len(layout)]); err == nil {
				return !t.Before(f.Cutoff)
			}
		}
	}
	return true
}

// Apply reads lines from in, forwards those that pass Allow, and closes out
// when in is closed or ctx is done.
func (f *Filter) Apply(ctx interface{ Done() <-chan struct{} }, in <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-in:
				if !ok {
					return
				}
				if f.Allow(line) {
					select {
					case out <- line:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out
}
