// Package multiline coalesces continuation lines into a single LogLine.
// A new logical line begins whenever the incoming text matches the
// configured start pattern; lines that do NOT match the start pattern
// are appended to the current buffer.
package multiline

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/user/logdrift/internal/runner"
)

// Joiner holds compiled state for multiline joining.
type Joiner struct {
	start   *regexp.Regexp
	timeout time.Duration
}

// New creates a Joiner. startPattern is a regex; a line matching it
// begins a new logical record. timeout flushes a pending record even
// if no new start line arrives.
func New(startPattern string, timeout time.Duration) (*Joiner, error) {
	re, err := regexp.Compile(startPattern)
	if err != nil {
		return nil, err
	}
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &Joiner{start: re, timeout: timeout}, nil
}

// Apply reads lines from in and writes coalesced lines to the returned channel.
func (j *Joiner) Apply(ctx context.Context, in <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		var buf []string
		var cur runner.LogLine
		flush := func() {
			if len(buf) == 0 {
				return
			}
			cur.Text = strings.Join(buf, "\n")
			select {
			case out <- cur:
			case <-ctx.Done():
			}
			buf = buf[:0]
		}
		timer := time.NewTimer(j.timeout)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				flush()
			case line, ok := <-in:
				if !ok {
					flush()
					return
				}
				if j.start.MatchString(line.Text) {
					flush()
					cur = line
					timer.Reset(j.timeout)
				}
				buf = append(buf, line.Text)
			}
		}
	}()
	return out
}
