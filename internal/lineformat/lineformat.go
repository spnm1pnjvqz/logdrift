// Package lineformat formats log lines using a user-defined template.
// Supported placeholders: {service}, {text}, {time}.
package lineformat

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/user/logdrift/internal/runner"
)

// Formatter replaces placeholders in a template with log line fields.
type Formatter struct {
	template string
}

// New returns a Formatter for the given template.
// Returns an error if the template contains no recognised placeholder.
func New(template string) (*Formatter, error) {
	if template == "" {
		return nil, fmt.Errorf("lineformat: template must not be empty")
	}
	known := []string{"{service}", "{text}", "{time}"}
	found := false
	for _, p := range known {
		if strings.Contains(template, p) {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("lineformat: template must contain at least one of %v", known)
	}
	return &Formatter{template: template}, nil
}

// Format applies the template to a single log line.
func (f *Formatter) Format(line runner.LogLine) string {
	out := f.template
	out = strings.ReplaceAll(out, "{service}", line.Service)
	out = strings.ReplaceAll(out, "{text}", line.Text)
	out = strings.ReplaceAll(out, "{time}", time.Now().UTC().Format(time.RFC3339))
	return out
}

// Apply reads lines from in, formats each one, and emits the result.
// The output channel is closed when in is closed or ctx is done.
func Apply(ctx context.Context, f *Formatter, in <-chan runner.LogLine) <-chan runner.LogLine {
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
				line.Text = f.Format(line)
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
