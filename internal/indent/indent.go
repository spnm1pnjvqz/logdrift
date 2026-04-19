// Package indent prepends a fixed string to every log line's text,
// useful for visually nesting output from child services.
package indent

import (
	"context"
	"errors"
	"fmt"

	"github.com/celrenshaw/logdrift/internal/runner"
)

// Indenter prepends a prefix string to each log line's text.
type Indenter struct {
	prefix string
}

// New returns an Indenter that prepends prefix to every line.
// prefix must be non-empty.
func New(prefix string) (*Indenter, error) {
	if prefix == "" {
		return nil, errors.New("indent: prefix must not be empty")
	}
	return &Indenter{prefix: prefix}, nil
}

// Stamp returns a copy of l with the text indented.
func (in *Indenter) Stamp(l runner.LogLine) runner.LogLine {
	l.Text = fmt.Sprintf("%s%s", in.prefix, l.Text)
	return l
}

// Apply reads lines from src, stamps each one, and forwards them to the
// returned channel. The channel is closed when src is closed or ctx is
// cancelled.
func (in *Indenter) Apply(ctx context.Context, src <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case l, ok := <-src:
				if !ok {
					return
				}
				select {
				case out <- in.Stamp(l):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
