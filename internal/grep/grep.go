// Package grep provides real-time pattern matching across log lines,
// emitting only lines that match at least one pattern.
package grep

import (
	"fmt"
	"regexp"

	"github.com/celrenheit/logdrift/internal/runner"
)

// Grep filters log lines by compiled regular expressions.
type Grep struct {
	patterns []*regexp.Regexp
	invert   bool
}

// New creates a Grep filter. If invert is true, lines matching any pattern
// are excluded rather than included.
func New(patterns []string, invert bool) (*Grep, error) {
	if len(patterns) == 0 {
		return nil, fmt.Errorf("grep: at least one pattern required")
	}
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("grep: invalid pattern %q: %w", p, err)
		}
		compiled = append(compiled, re)
	}
	return &Grep{patterns: compiled, invert: invert}, nil
}

// Match reports whether the line text matches any pattern.
func (g *Grep) Match(text string) bool {
	for _, re := range g.patterns {
		if re.MatchString(text) {
			return !g.invert
		}
	}
	return g.invert
}

// Apply reads from in, forwarding only lines that satisfy Match, and closes
// the returned channel when in is closed or ctx is done.
func (g *Grep) Apply(ctx interface{ Done() <-chan struct{} }, in <-chan runner.LogLine) <-chan runner.LogLine {
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
				if g.Match(line.Text) {
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
