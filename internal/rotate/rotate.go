// Package rotate detects log file rotation and re-opens tailed files.
package rotate

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

// DefaultPollInterval is how often the watcher checks for rotation.
const DefaultPollInterval = 500 * time.Millisecond

// Event is emitted when a rotation is detected for a named source.
type Event struct {
	Source string
	Path   string
}

// Watcher polls a set of file paths and emits an Event whenever a file
// is rotated (inode changes or file shrinks).
type Watcher struct {
	sources      map[string]string // name -> path
	pollInterval time.Duration
	inodes        map[string]uint64
	sizes         map[string]int64
}

// New creates a Watcher for the given name->path mapping.
func New(sources map[string]string, interval time.Duration) *Watcher {
	if interval <= 0 {
		interval = DefaultPollInterval
	}
	return &Watcher{
		sources:      sources,
		pollInterval: interval,
		inodes:        make(map[string]uint64),
		sizes:         make(map[string]int64),
	}
}

// Watch starts polling and sends rotation Events on the returned channel.
// The channel is closed when ctx is cancelled.
func (w *Watcher) Watch(ctx context.Context) <-chan Event {
	ch := make(chan Event, len(w.sources))
	go func() {
		defer close(ch)
		ticker := time.NewTicker(w.pollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for name, path := range w.sources {
					if ev, ok := w.check(name, path); ok {
						log.Debug().Str("source", name).Str("path", path).n						select {
						case ch <- ev:
						case <-ctx.Done():
			\n						}
					}
				}
			}
		}
	}()
	return ch
}

func (w *Watcher) check(name, path string) (Event, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return Event{}, false
	}
	i inoOf(info)
	size := info.Size()
	prev := w.inodes[name]
	if !seen {
		w.inodes[name] = inode
		w.sizes[name] = size
		return Event{}, false
	}
	rotated := inode != prev || size < w.sizes[name]
	w.inodes[name] = inode
	w.sizes[name] = size
	if rotated {
		return Event{Source: name, Path: path}, true
	}
	return Event{}, false
}
