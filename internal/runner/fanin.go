package runner

import (
	"context"
	"sync"
)

// FanIn merges multiple Line channels into a single channel.
// The returned channel is closed once all input channels are drained.
func FanIn(ctx context.Context, channels ...<-chan Line) <-chan Line {
	out := make(chan Line, 128)
	var wg sync.WaitGroup

	forward := func(ch <-chan Line) {
		defer wg.Done()
		for {
			select {
			case line, ok := <-ch:
				if !ok {
					return
				}
				select {
				case out <- line:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}

	wg.Add(len(channels))
	for _, ch := range channels {
		go forward(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
