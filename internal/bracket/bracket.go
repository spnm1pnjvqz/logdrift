// Package bracket wraps each log line's text with a configurable open and
// close string, e.g. "[" and "]" or "<" and ">".  It is useful when
// piping logdrift output into tools that expect delimited records.
package bracket

import (
	"context"
	"fmt"

	"github.com/user/logdrift/internal/runner"
)

// Bracketer prepends and appends fixed strings to every log line.
type Bracketer struct {
	open  string
	close string
}

// New returns a Bracketer or an error when both open and close are empty.
func New(open, close string) (*Bracketer, error) {
	if open == "" && close == "" {
		return nil, fmt.Errorf("bracket: at least one of open or close must be non-empty")
	}
	return &Bracketer{open: open, close: close}, nil
}

// Stamp returns the line text wrapped with the configured delimiters.
func (b *Bracketer) Stamp(line runner.LogLine) runner.LogLine {
	line.Text = b.open + line.Text + b.close
	return line
}

// Apply reads from src, wraps every line, and forwards it to the returned
// channel.  The output channel is closed when src is drained or ctx is
// cancelled.
func (b *Bracketer) Apply(ctx context.Context, src <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
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
				select {
				case out <- b.Stamp(line):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
