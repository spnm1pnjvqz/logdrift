// Package merge provides utilities for combining multiple log-line channels
// into a single ordered stream, optionally interleaving by timestamp.
package merge

import (
	"context"
	"sync"

	"github.com/logdrift/logdrift/internal/runner"
)

// Merge fans in multiple LogLine channels into one. Lines are emitted in
// arrival order (non-deterministic across sources). The output channel is
// closed once all input channels are drained or ctx is cancelled.
func Merge(ctx context.Context, sources ...<-chan runner.LogLine) (<-chan runner.LogLine, error) {
	if len(sources) == 0 {
		return nil, ErrNoSources
	}

	out := make(chan runner.LogLine, len(sources)*8)
	var wg sync.WaitGroup

	for _, src := range sources {
		wg.Add(1)
		go func(ch <-chan runner.LogLine) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case line, ok := <-ch:
					if !ok {
						return
					}
					select {
					case out <- line:
					case <-ctx.Done():
						return
					}
				}
			}
		}(src)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out, nil
}
