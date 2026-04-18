// Package strip removes ANSI escape codes and/or leading/trailing
// whitespace from log lines before further processing.
package strip

import (
	"context"
	"regexp"
	"strings"

	"github.com/user/logdrift/internal/runner"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// Options controls what Strip removes.
type Options struct {
	ANSI       bool
	Whitespace bool
}

// Stripper applies stripping rules to log line text.
type Stripper struct {
	opts Options
}

// New returns a Stripper for the given Options.
// At least one option must be enabled.
func New(opts Options) (*Stripper, error) {
	if !opts.ANSI && !opts.Whitespace {
		return nil, errNoOptions
	}
	return &Stripper{opts: opts}, nil
}

// Apply returns a cleaned copy of text.
func (s *Stripper) Apply(text string) string {
	if s.opts.ANSI {
		text = ansiEscape.ReplaceAllString(text, "")
	}
	if s.opts.Whitespace {
		text = strings.TrimSpace(text)
	}
	return text
}

// Stream reads lines from in, strips each one, and sends results to the
// returned channel. The channel is closed when in is closed or ctx is done.
func (s *Stripper) Stream(ctx context.Context, in <-chan runner.LogLine) <-chan runner.LogLine {
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
				line.Text = s.Apply(line.Text)
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
