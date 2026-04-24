// Package splitter routes log lines to named output channels based on
// per-service routing rules defined by regular-expression patterns.
package splitter

import (
	"context"
	"fmt"
	"regexp"

	"github.com/user/logdrift/internal/runner"
)

// Rule maps a compiled pattern to a named bucket.
type Rule struct {
	Bucket  string
	pattern *regexp.Regexp
}

// Splitter routes lines into named output channels.
type Splitter struct {
	rules   []Rule
	default_ string
}

// New creates a Splitter. rawRules maps bucket names to regex patterns.
// defaultBucket receives lines that match no rule; it may be empty to drop
// unmatched lines.
func New(rawRules map[string]string, defaultBucket string) (*Splitter, error) {
	if len(rawRules) == 0 {
		return nil, fmt.Errorf("splitter: at least one rule is required")
	}
	rules := make([]Rule, 0, len(rawRules))
	for bucket, pat := range rawRules {
		if bucket == "" {
			return nil, fmt.Errorf("splitter: bucket name must not be empty")
		}
		re, err := regexp.Compile(pat)
		if err != nil {
			return nil, fmt.Errorf("splitter: invalid pattern for bucket %q: %w", bucket, err)
		}
		rules = append(rules, Rule{Bucket: bucket, pattern: re})
	}
	return &Splitter{rules: rules, default_: defaultBucket}, nil
}

// Route returns the bucket name for line text, or the default bucket if no
// rule matches. An empty string means the line should be dropped.
func (s *Splitter) Route(text string) string {
	for _, r := range s.rules {
		if r.pattern.MatchString(text) {
			return r.Bucket
		}
	}
	return s.default_
}

// Apply reads lines from src, routes each one, and forwards it to the
// appropriate output channel. Outputs is a map of bucket -> channel that the
// caller must supply (and may create via MakeOutputs).
func (s *Splitter) Apply(
	ctx context.Context,
	src <-chan runner.LogLine,
	outputs map[string]chan runner.LogLine,
) {
	go func() {
		defer func() {
			for _, ch := range outputs {
				close(ch)
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-src:
				if !ok {
					return
				}
				bucket := s.Route(line.Text)
				if ch, found := outputs[bucket]; found {
					select {
					case ch <- line:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
}
