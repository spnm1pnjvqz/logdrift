package tail

import (
	"context"
	"fmt"

	"github.com/user/logdrift/internal/runner"
)

// FileSource describes a file to tail with an associated service name.
type FileSource struct {
	Service string
	Path    string
}

// TailAll starts a Tailer for each FileSource and fans all output into
// a single merged channel. The channel is closed once all tailers stop.
func TailAll(ctx context.Context, sources []FileSource) (<-chan runner.LogLine, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("tail: no sources provided")
	}

	channels := make([]<-chan runner.LogLine, 0, len(sources))
	for _, src := range sources {
		tr, err := New(src.Path)
		if err != nil {
			return nil, fmt.Errorf("tail: %s: %w", src.Service, err)
		}
		ch, err := tr.Tail(ctx, src.Service)
		if err != nil {
			return nil, fmt.Errorf("tail: %s: %w", src.Service, err)
		}
		channels = append(channels, ch)
	}

	return mergeChans(ctx, channels), nil
}

// mergeChans fans-in multiple LogLine channels into one.
func mergeChans(ctx context.Context, chans []<-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine, 64)
	done := make(chan struct{}, len(chans))

	for _, ch := range chans {
		go func(c <-chan runner.LogLine) {
			defer func() { done <- struct{}{} }()
			for {
				select {
				case <-ctx.Done():
					return
				case line, ok := <-c:
					if !ok {
						return
					}
					out <- line
				}
			}
		}(ch)
	}

	go func() {
		for i := 0; i < len(chans); i++ {
			<-done
		}
		close(out)
	}()

	return out
}
