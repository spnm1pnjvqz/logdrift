// Package top tracks the most frequently seen log lines per service
// within a sliding count window.
package top

import (
	"errors"
	"sort"
	"sync"

	"github.com/user/logdrift/internal/runner"
)

// Entry holds a log line and its occurrence count.
type Entry struct {
	Line  string
	Count int
}

// Tracker counts line frequencies per service and returns the top-N.
type Tracker struct {
	mu      sync.Mutex
	n       int
	counts  map[string]map[string]int // service -> line -> count
}

// New creates a Tracker that will return the top n lines per service.
func New(n int) (*Tracker, error) {
	if n <= 0 {
		return nil, errors.New("top: n must be greater than zero")
	}
	return &Tracker{
		n:      n,
		counts: make(map[string]map[string]int),
	}, nil
}

// Add records a log line for the given service.
func (t *Tracker) Add(service, line string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.counts[service]; !ok {
		t.counts[service] = make(map[string]int)
	}
	t.counts[service][line]++
}

// Top returns the top-n entries for a service, sorted descending by count.
func (t *Tracker) Top(service string) []Entry {
	t.mu.Lock()
	defer t.mu.Unlock()
	lines := t.counts[service]
	entries := make([]Entry, 0, len(lines))
	for l, c := range lines {
		entries = append(entries, Entry{Line: l, Count: c})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Count > entries[j].Count
	})
	if len(entries) > t.n {
		entries = entries[:t.n]
	}
	return entries
}

// Apply consumes lines from ch, records them, and forwards them unchanged.
func (t *Tracker) Apply(ctx interface{ Done() <-chan struct{} }, ch <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-ch:
				if !ok {
					return
				}
				t.Add(line.Service, line.Text)
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
