// Package jsonformat pretty-prints JSON log lines, leaving non-JSON lines unchanged.
package jsonformat

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/user/logdrift/internal/runner"
)

// Formatter pretty-prints JSON lines.
type Formatter struct {
	indent string
}

// New returns a Formatter. indent is the string used for each indentation level.
// If indent is empty, "  " (two spaces) is used.
func New(indent string) *Formatter {
	if indent == "" {
		indent = "  "
	}
	return &Formatter{indent: indent}
}

// Format attempts to pretty-print the text of l. If the text is not valid JSON
// the original text is returned unchanged.
func (f *Formatter) Format(l runner.LogLine) runner.LogLine {
	trimmed := bytes.TrimSpace([]byte(l.Text))
	if len(trimmed) == 0 || trimmed[0] != '{' && trimmed[0] != '[' {
		return l
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, trimmed, "", f.indent); err != nil {
		return l
	}
	l.Text = buf.String()
	return l
}

// Apply reads lines from in, formats each one, and sends results to the
// returned channel. The channel is closed when in is closed or ctx is done.
func (f *Formatter) Apply(ctx context.Context, in <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case l, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- f.Format(l):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
