// Package normalize transforms log line text by applying
// case folding, whitespace collapsing, and optional trimming
// so that downstream comparisons are more robust.
package normalize

import (
	"errors"
	"regexp"
	"strings"

	"github.com/myorg/logdrift/internal/runner"
)

var multiSpace = regexp.MustCompile(`\s+`)

// Options controls which normalizations are applied.
type Options struct {
	// Lowercase folds all text to lower case.
	Lowercase bool
	// CollapseSpaces replaces runs of whitespace with a single space.
	CollapseSpaces bool
	// Trim removes leading and trailing whitespace.
	Trim bool
}

// Normalizer applies text normalizations to log lines.
type Normalizer struct {
	opts Options
}

// New returns a Normalizer configured with opts.
// Returns an error if no normalization option is enabled.
func New(opts Options) (*Normalizer, error) {
	if !opts.Lowercase && !opts.CollapseSpaces && !opts.Trim {
		return nil, errors.New("normalize: at least one option must be enabled")
	}
	return &Normalizer{opts: opts}, nil
}

// Apply returns a normalized copy of the given log line.
func (n *Normalizer) Apply(line runner.LogLine) runner.LogLine {
	text := line.Text
	if n.opts.Trim {
		text = strings.TrimSpace(text)
	}
	if n.opts.CollapseSpaces {
		text = multiSpace.ReplaceAllString(text, " ")
	}
	if n.opts.Lowercase {
		text = strings.ToLower(text)
	}
	line.Text = text
	return line
}

// ApplyAll reads from in, normalizes each line, and sends results to the
// returned channel. The output channel is closed when in is closed or ctx
// is cancelled.
func (n *Normalizer) ApplyAll(ctx interface{ Done() <-chan struct{} }, in <-chan runner.LogLine) <-chan runner.LogLine {
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
				select {
				case out <- n.Apply(line):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
