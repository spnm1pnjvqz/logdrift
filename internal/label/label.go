// Package label attaches a fixed service label to every log line that passes
// through the pipeline, overwriting any previously set label.
package label

import (
	"context"
	"fmt"

	"github.com/user/logdrift/internal/runner"
)

// Labeler stamps every incoming LogLine with a fixed service name.
type Labeler struct {
	service string
}

// New returns a Labeler for the given service name.
func New(service string) (*Labeler, error) {
	if service == "" {
		return nil, fmt.Errorf("label: service name must not be empty")
	}
	return &Labeler{service: service}, nil
}

// Apply reads lines from in, sets the Service field, and forwards them to the
// returned channel. The output channel is closed when ctx is cancelled or in
// is closed.
func (l *Labeler) Apply(ctx context.Context, in <-chan runner.LogLine) <-chan runner.LogLine {
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
				line.Service = l.service
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
