// Package mask replaces substrings matching patterns with a fixed placeholder,
// useful for hiding secrets or PII in log lines before display or storage.
package mask

import (
	"fmt"
	"regexp"

	"github.com/yourorg/logdrift/internal/runner"
)

const defaultPlaceholder = "[MASKED]"

// Rule pairs a compiled pattern with its replacement string.
type Rule struct {
	re          *regexp.Regexp
	replacement string
}

// Masker holds a set of masking rules.
type Masker struct {
	rules []Rule
}

// New compiles each pattern into a Rule. placeholder is used for every match;
// pass an empty string to use the default "[MASKED]".
func New(patterns []string, placeholder string) (*Masker, error) {
	if len(patterns) == 0 {
		return nil, fmt.Errorf("mask: at least one pattern required")
	}
	if placeholder == "" {
		placeholder = defaultPlaceholder
	}
	rules := make([]Rule, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("mask: invalid pattern %q: %w", p, err)
		}
		rules = append(rules, Rule{re: re, replacement: placeholder})
	}
	return &Masker{rules: rules}, nil
}

// Apply returns a copy of text with all rule matches replaced.
func (m *Masker) Apply(text string) string {
	for _, r := range m.rules {
		text = r.re.ReplaceAllString(text, r.replacement)
	}
	return text
}

// Transform reads lines from in, masks each line's Text, and forwards to the
// returned channel. The channel is closed when in is closed or ctx is done.
func (m *Masker) Transform(ctx interface{ Done() <-chan struct{} }, in <-chan runner.LogLine) <-chan runner.LogLine {
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
				line.Text = m.Apply(line.Text)
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
