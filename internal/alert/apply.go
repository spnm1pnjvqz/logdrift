package alert

import (
	"context"

	"github.com/user/logdrift/internal/runner"
)

// Apply reads LogLines from in, checks each against the Alerter, and forwards
// matching events to the returned channel. The output channel is closed when
// ctx is cancelled or in is closed.
func Apply(ctx context.Context, a *Alerter, in <-chan runner.LogLine) <-chan Event {
	out := make(chan Event)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case ll, ok := <-in:
				if !ok {
					return
				}
				for _, ev := range a.Check(ll.Service, ll.Text) {
					select {
					case out <- ev:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out
}
