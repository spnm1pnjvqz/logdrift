// Package prefix prepends a fixed string to every log line's text.
package prefix

import (
	"context"
	"fmt"

	"github.com/yourorg/logdrift/internal/runner"
)

// Prefixer prepends a string to each log line.
type Prefixer struct {
	prefix string
}

// New returns a Prefixer that prepends p to every line.
// Returns an error if p is empty.
func New(p string) (*Prefixer, error) {
	if p == "" {
		return nil, fmt.Errorf("prefix: prefix string must not be empty")
	}
	return &Prefixer{prefix: p}, nil
}

// Apply reads lines from in, prepends the prefix to each line's Text,
// and sends the result to the returned channel.
// The output channel is closed when in is closed or ctx is cancelled.
func (p *Prefixer) Apply(ctx context.Context, in <-chan runner.LogLine) <-chan runner.LogLine {
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
				line.Text = p.prefix + line.Text
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
