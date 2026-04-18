// Package sequence assigns a monotonically increasing counter to each log line
// so consumers can detect gaps or reordering across merged streams.
package sequence

import (
	"fmt"
	"sync/atomic"

	"github.com/user/logdrift/internal/runner"
)

// Sequencer stamps log lines with a global sequence number.
type Sequencer struct {
	counter uint64
	prefix  string
}

// New returns a Sequencer. prefix is prepended to the counter tag, e.g. "#".
func New(prefix string) (*Sequencer, error) {
	if prefix == "" {
		prefix = "#"
	}
	return &Sequencer{prefix: prefix}, nil
}

// Stamp appends a sequence tag to the line text and returns the updated line.
func (s *Sequencer) Stamp(line runner.LogLine) runner.LogLine {
	n := atomic.AddUint64(&s.counter, 1)
	line.Text = fmt.Sprintf("%s [%s%d]", line.Text, s.prefix, n)
	return line
}

// Apply reads from in, stamps each line, and writes to the returned channel.
// The output channel is closed when in is closed or ctx is done.
func (s *Sequencer) Apply(ctx interface{ Done() <-chan struct{} }, in <-chan runner.LogLine) <-chan runner.LogLine {
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
				select {
				case out <- s.Stamp(line):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
