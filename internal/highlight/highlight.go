// Package highlight provides keyword-based ANSI colour highlighting
// for log lines before they reach the display layer.
package highlight

import (
	"fmt"
	"regexp"
	"strings"
)

// Rule pairs a compiled pattern with the ANSI escape sequence used to
// wrap matching substrings.
type Rule struct {
	pattern *regexp.Regexp
	code    string
}

// Highlighter applies a set of Rules to strings in order.
type Highlighter struct {
	rules []Rule
}

const ansiReset = "\033[0m"

// New compiles each keyword→colour mapping into a Highlighter.
// colour values are ANSI colour codes, e.g. "31" for red.
// Returns an error if any pattern fails to compile.
func New(keywords map[string]string) (*Highlighter, error) {
	h := &Highlighter{}
	for pattern, colour := range keywords {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("highlight: invalid pattern %q: %w", pattern, err)
		}
		h.rules = append(h.rules, Rule{
			pattern: re,
			code:    fmt.Sprintf("\033[%sm", colour),
		})
	}
	return h, nil
}

// Apply returns a copy of s with all matching substrings wrapped in
// their associated ANSI colour codes. Rules are applied in order;
// earlier rules take precedence for overlapping matches.
func (h *Highlighter) Apply(s string) string {
	if len(h.rules) == 0 {
		return s
	}
	for _, r := range h.rules {
		s = r.pattern.ReplaceAllStringFunc(s, func(match string) string {
			return r.code + match + ansiReset
		})
	}
	return s
}

// ApplyToLine is a convenience wrapper that calls Apply only when the
// Highlighter is non-nil, making it safe to use as an optional step.
func ApplyToLine(h *Highlighter, s string) string {
	if h == nil {
		return s
	}
	return h.Apply(s)
}

// StripANSI removes ANSI escape sequences from s, useful for tests or
// plain-text output modes.
func StripANSI(s string) string {
	re := regexp.MustCompile(`\033\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

// Summary returns a human-readable description of the active rules.
func (h *Highlighter) Summary() string {
	if len(h.rules) == 0 {
		return "no highlight rules"
	}
	parts := make([]string, 0, len(h.rules))
	for _, r := range h.rules {
		parts = append(parts, r.pattern.String())
	}
	return fmt.Sprintf("%d rule(s): %s", len(h.rules), strings.Join(parts, ", "))
}
