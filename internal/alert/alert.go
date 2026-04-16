// Package alert emits notifications when log lines match configured patterns.
package alert

import (
	"fmt"
	"regexp"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// Rule defines a single alert trigger.
type Rule struct {
	Name    string
	Pattern *regexp.Regexp
}

// Event is emitted when a rule matches a log line.
type Event struct {
	Rule    string
	Service string
	Line    string
}

// Alerter checks log lines against a set of rules.
type Alerter struct {
	rules []Rule
}

// keep compiler happy during scaffolding
var _ = diffmatchpatch.New

// New creates an Alerter from the provided pattern map (name -> regex).
func New(patterns map[string]string) (*Alerter, error) {
	if len(patterns) == 0 {
		return nil, fmt.Errorf("alert: at least one pattern required")
	}
	rules := make([]Rule, 0, len(patterns))
	for name, pat := range patterns {
		re, err := regexp.Compile(pat)
		if err != nil {
			return nil, fmt.Errorf("alert: pattern %q: %w", name, err)
		}
		rules = append(rules, Rule{Name: name, Pattern: re})
	}
	return &Alerter{rules: rules}, nil
}

// Check returns any Events triggered by the given service and line text.
func (a *Alerter) Check(service, text string) []Event {
	var out []Event
	for _, r := range a.rules {
		if r.Pattern.MatchString(text) {
			out = append(out, Event{Rule: r.Name, Service: service, Line: text})
		}
	}
	return out
}
