// Package fork fans a single log-line channel out to N independent channels.
package fork

import (
	"context"
	"sync"

	"github.com/logdrift/logdrift/internal/runner"
)

// Fork duplicates every line from src into n independent output channels.
// Each output channel is buffered with the given capacity.
// All outputs are closed once src is drained or ctx is cancelled.
func Fork(ctx context.Context, src <-chan runner.LogLine, n int, buf int) []<-chan runner.LogLine {
	outs := make([]chan runner.LogLine, n)
	result := make([]<-chan runner.LogLine, n)
	for i := 0; i < n; i++ {
		ch := make(chan runner.LogLine, buf)
		outs[i] = ch
		result[i] = ch
	}

	go func() {
		defer func() {
			for _, ch := range outs {
				close(ch)
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-src:
				if !ok {
					return
				}
				var wg sync.WaitGroup
				for _, ch := range outs {
					wg.Add(1)
					go func(c chan runner.LogLine) {
						defer wg.Done()
						select {
						case c <- line:
						case <-ctx.Done():
						}
					}(ch)
				}
				wg.Wait()
			}
		}
	}()

	return result
}
