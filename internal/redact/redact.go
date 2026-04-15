// Package redact provides pattern-based redaction of sensitive data in log lines.
package redact

import (
	"fmt"
	"regexp"
)

// Rule holds a compiled pattern and its replacement string.
type Rule struct {
	pattern     *regexp.Regexp
	replacement string
}

// Redactor applies a set of redaction rules to log lines.
type Redactor struct {
	rules []Rule
}

// New compiles the provided patterns into a Redactor.
// Each entry in patterns maps a regex string to a replacement string.
// Returns an error if any pattern fails to compile.
func New(patterns map[string]string) (*Redactor, error) {
	rules := make([]Rule, 0, len(patterns))
	for pat, repl := range patterns {
		re, err := regexp.Compile(pat)
		if err != nil {
			return nil, fmt.Errorf("redact: invalid pattern %q: %w", pat, err)
		}
		rules = append(rules, Rule{pattern: re, replacement: repl})
	}
	return &Redactor{rules: rules}, nil
}

// Apply returns a copy of s with all matching patterns replaced.
func (r *Redactor) Apply(s string) string {
	for _, rule := range r.rules {
		s = rule.pattern.ReplaceAllString(s, rule.replacement)
	}
	return s
}

// ApplyToChannel reads lines from in, applies redaction, and sends results to
// the returned channel. The output channel is closed when in is closed.
func (r *Redactor) ApplyToChannel(in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for line := range in {
			out <- r.Apply(line)
		}
	}()
	return out
}

// RuleCount returns the number of redaction rules in the Redactor.
func (r *Redactor) RuleCount() int {
	return len(r.rules)
}

// AddRule compiles and appends a single new redaction rule to the Redactor.
// Returns an error if the pattern fails to compile.
func (r *Redactor) AddRule(pattern, replacement string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("redact: invalid pattern %q: %w", pattern, err)
	}
	r.rules = append(r.rules, Rule{pattern: re, replacement: replacement})
	return nil
}
