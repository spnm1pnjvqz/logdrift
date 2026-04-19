// Package columns formats log lines into fixed-width aligned columns
// by splitting on a configurable delimiter.
package columns

import (
	"fmt"
	"strings"

	"github.com/user/logdrift/internal/runner"
)

// Formatter aligns fields in log lines into fixed-width columns.
type Formatter struct {
	delimiter string
	widths    []int
}

// New returns a Formatter that splits on delim and pads each field to the
// corresponding width in widths. If a line has more fields than widths entries
// the remaining fields are appended unstyled.
func New(delim string, widths []int) (*Formatter, error) {
	if delim == "" {
		return nil, fmt.Errorf("columns: delimiter must not be empty")
	}
	if len(widths) == 0 {
		return nil, fmt.Errorf("columns: at least one column width required")
	}
	for i, w := range widths {
		if w <= 0 {
			return nil, fmt.Errorf("columns: width[%d] must be positive, got %d", i, w)
		}
	}
	return &Formatter{delimiter: delim, widths: widths}, nil
}

// Format aligns the text of a LogLine and returns the updated line.
func (f *Formatter) Format(line runner.LogLine) runner.LogLine {
	parts := strings.Split(line.Text, f.delimiter)
	var b strings.Builder
	for i, p := range parts {
		if i < len(f.widths) {
			b.WriteString(fmt.Sprintf("%-*s", f.widths[i], p))
		} else {
			if i > 0 {
				b.WriteString(f.delimiter)
			}
			b.WriteString(p)
		}
	}
	line.Text = b.String()
	return line
}

// Apply reads from src, formats each line, and sends it to the returned channel.
// The output channel is closed when src is closed or ctx is done.
func (f *Formatter) Apply(src <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		for line := range src {
			out <- f.Format(line)
		}
	}()
	return out
}
