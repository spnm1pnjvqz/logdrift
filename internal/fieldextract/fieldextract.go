// Package fieldextract parses structured key=value or JSON fields from log lines.
package fieldextract

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/user/logdrift/internal/runner"
)

var kvRe = regexp.MustCompile(`(\w+)=("[^"]*"|\S+)`)

// Extractor pulls named fields from log line text.
type Extractor struct {
	fields []string
	useJSON bool
}

// New creates an Extractor for the given field names.
// If json is true, lines are parsed as JSON objects; otherwise key=value.
func New(fields []string, useJSON bool) (*Extractor, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("fieldextract: at least one field required")
	}
	return &Extractor{fields: fields, useJSON: useJSON}, nil
}

// Extract returns a map of requested field values found in text.
func (e *Extractor) Extract(text string) map[string]string {
	if e.useJSON {
		return e.extractJSON(text)
	}
	return e.extractKV(text)
}

func (e *Extractor) extractKV(text string) map[string]string {
	want := make(map[string]bool, len(e.fields))
	for _, f := range e.fields {
		want[f] = true
	}
	out := make(map[string]string)
	for _, m := range kvRe.FindAllStringSubmatch(text, -1) {
		if want[m[1]] {
			out[m[1]] = strings.Trim(m[2], `"`)
		}
	}
	return out
}

func (e *Extractor) extractJSON(text string) map[string]string {
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(text), &obj); err != nil {
		return map[string]string{}
	}
	out := make(map[string]string)
	for _, f := range e.fields {
		if v, ok := obj[f]; ok {
			out[f] = fmt.Sprintf("%v", v)
		}
	}
	return out
}

// Apply reads lines from in, annotates each line's Text with extracted fields
// prepended, and sends to the returned channel. Closes when ctx done or in closed.
func (e *Extractor) Apply(in <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		for line := range in {
			fields := e.Extract(line.Text)
			if len(fields) > 0 {
				parts := make([]string, 0, len(fields)+1)
				for k, v := range fields {
					parts = append(parts, k+"="+v)
				}
				parts = append(parts, line.Text)
				line.Text = strings.Join(parts, " ")
			}
			out <- line
		}
	}()
	return out
}
