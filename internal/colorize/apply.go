package colorize

import (
	"context"

	"github.com/iamcathal/logdrift/internal/runner"
)

// Apply reads LogLines from in, wraps each line's Text with the service color,
// and forwards the modified line to the returned channel.
// The output channel is closed when ctx is cancelled or in is closed.
func Apply(ctx context.Context, c *Colorizer, in <-chan runner.LogLine) <-chan runner.LogLine {
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
				line.Text = c.Wrap(line.Service, line.Text)
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
