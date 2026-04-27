// Package tee duplicates a log line stream into two independent output channels.
// It is useful when you want to apply two different processing pipelines
// to the same source without consuming the channel twice.
package tee

import (
	"context"

	"github.com/user/logdrift/internal/runner"
)

// Tee reads from src and writes every line to both out1 and out2.
// Both output channels are closed when src is exhausted or ctx is cancelled.
// If a receiver is slow, Tee will block until both outputs have accepted the
// line, preserving ordering guarantees.
func Tee(ctx context.Context, src <-chan runner.LogLine) (<-chan runner.LogLine, <-chan runner.LogLine) {
	out1 := make(chan runner.LogLine)
	out2 := make(chan runner.LogLine)

	go func() {
		defer close(out1)
		defer close(out2)

		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-src:
				if !ok {
					return
				}
				// Write to both outputs; block until each accepts.
				select {
				case out1 <- line:
				case <-ctx.Done():
					return
				}
				select {
				case out2 <- line:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out1, out2
}
