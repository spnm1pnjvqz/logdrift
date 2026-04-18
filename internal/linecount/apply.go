package linecount

import (
	"context"

	"github.com/yourorg/logdrift/internal/runner"
)

// Apply reads log lines from in, increments the counter for each line's
// service, and forwards every line unchanged to the returned channel.
func (c *Counter) Apply(ctx context.Context, in <-chan runner.LogLine) <-chan runner.LogLine {
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
				// best-effort increment; ignore empty service names
				_ = c.Add(line.Service)
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
