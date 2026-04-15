package filter

import "github.com/yourorg/logdrift/internal/runner"

// Apply reads LogLines from in, passes each through f, and forwards
// matching lines to the returned channel. The output channel is closed
// when in is closed.
func Apply(f *Filter, in <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		for ll := range in {
			if f.Allow(ll.Line) {
				out <- ll
			}
		}
	}()
	return out
}
