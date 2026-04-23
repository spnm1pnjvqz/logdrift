package tail

import (
	"context"
	"os"
	"time"

	"github.com/yourorg/logdrift/internal/rotate"
)

// WatchConfig holds configuration for a watched tail source.
type WatchConfig struct {
	Path     string
	Interval time.Duration
}

// WatchResult is emitted when a file rotation is detected and a new tailer
// has been started. The Lines channel carries lines from the new file.
type WatchResult struct {
	Path  string
	Lines <-chan string
}

// Watch tails a file and transparently restarts the tailer whenever a
// rotation event is detected (file shrinks or inode changes). It emits a
// WatchResult each time a new tailer is started, including the initial one.
// The returned channel is closed when ctx is cancelled.
func Watch(ctx context.Context, cfg WatchConfig) (<-chan WatchResult, error) {
	if _, err := os.Stat(cfg.Path); err != nil {
		return nil, err
	}

	interval := cfg.Interval
	if interval <= 0 {
		interval = time.Second
	}

	out := make(chan WatchResult, 1)

	go func() {
		defer close(out)

		for {
			// Start a fresh tailer for the current file.
			tailer, err := New(cfg.Path)
			if err != nil {
				return
			}

			tailCtx, cancelTail := context.WithCancel(ctx)
			lines := tailer.Start(tailCtx)

			select {
			case out <- WatchResult{Path: cfg.Path, Lines: lines}:
			case <-ctx.Done():
				cancelTail()
				return
			}

			// Watch for rotation events.
			watcher, werr := rotate.New(cfg.Path, rotate.Options{Interval: interval})
			if werr != nil {
				cancelTail()
				return
			}
			events := watcher.Watch(ctx)

			select {
			case <-events:
				// Rotation detected — cancel current tailer and restart.
				cancelTail()
			case <-ctx.Done():
				cancelTail()
				return
			}
		}
	}()

	return out, nil
}
