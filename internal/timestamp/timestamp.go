// Package timestamp provides utilities for parsing, formatting, and
// attaching timestamps to log lines as they flow through the pipeline.
package timestamp

import (
	"fmt"
	"time"

	"github.com/user/logdrift/internal/runner"
)

// Format controls how timestamps are rendered when prepended to log lines.
type Format string

const (
	FormatRFC3339  Format = "rfc3339"
	FormatUnix     Format = "unix"
	FormatKitchen  Format = "kitchen"
	FormatRelative Format = "relative"
)

// Stamper prepends timestamps to log lines.
type Stamper struct {
	format  Format
	started time.Time
}

// New creates a Stamper with the given format.
// Returns an error if the format is unrecognised.
func New(format Format) (*Stamper, error) {
	switch format {
	case FormatRFC3339, FormatUnix, FormatKitchen, FormatRelative:
		return &Stamper{format: format, started: time.Now()}, nil
	default:
		return nil, fmt.Errorf("timestamp: unknown format %q", format)
	}
}

// Stamp returns the log line with a timestamp prepended.
func (s *Stamper) Stamp(line runner.LogLine) runner.LogLine {
	ts := s.render(line.At)
	line.Text = ts + " " + line.Text
	return line
}

func (s *Stamper) render(t time.Time) string {
	switch s.format {
	case FormatRFC3339:
		return t.Format(time.RFC3339)
	case FormatUnix:
		return fmt.Sprintf("%d", t.Unix())
	case FormatKitchen:
		return t.Format(time.Kitchen)
	case FormatRelative:
		d := t.Sub(s.started).Truncate(time.Millisecond)
		return fmt.Sprintf("+%s", d)
	default:
		return t.Format(time.RFC3339)
	}
}

// Apply reads lines from in, stamps each one, and sends them to the returned channel.
// The output channel is closed when in is closed or ctx is done.
func (s *Stamper) Apply(in <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		for line := range in {
			out <- s.Stamp(line)
		}
	}()
	return out
}
