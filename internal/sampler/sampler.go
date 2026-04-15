// Package sampler provides rate-based sampling for log line channels.
// It allows keeping every Nth line from a stream, useful for reducing
// noise from high-throughput services without dropping all output.
package sampler

import (
	"fmt"

	"github.com/user/logdrift/internal/runner"
)

// Config holds sampler configuration.
type Config struct {
	// N specifies that every Nth line is kept. Must be >= 1.
	// A value of 1 means all lines are kept (no sampling).
	N int
}

// Validate returns an error if the Config is invalid.
func (c Config) Validate() error {
	if c.N < 1 {
		return fmt.Errorf("sampler: N must be >= 1, got %d", c.N)
	}
	return nil
}

// Apply reads lines from in, forwards every Nth line to the returned
// channel, and closes the output channel when in is closed.
// The caller is responsible for providing a valid Config (N >= 1).
func Apply(cfg Config, in <-chan runner.Line) <-chan runner.Line {
	out := make(chan runner.Line)
	go func() {
		defer close(out)
		var count int
		for line := range in {
			count++
			if count%cfg.N == 0 {
				out <- line
			}
		}
	}()
	return out
}
