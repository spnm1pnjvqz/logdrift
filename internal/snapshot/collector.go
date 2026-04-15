package snapshot

import (
	"context"

	"github.com/yourorg/logdrift/internal/runner"
)

// Collector drains a fan-in line channel and accumulates entries into a
// Snapshot until the context is cancelled or the channel is closed.
type Collector struct {
	snap *Snapshot
}

// NewCollector returns a Collector backed by a fresh Snapshot.
func NewCollector() *Collector {
	return &Collector{snap: New()}
}

// Collect reads from ch until it is closed or ctx is done, adding each
// LogLine to the internal snapshot. It returns the completed Snapshot.
func (c *Collector) Collect(ctx context.Context, ch <-chan runner.LogLine) *Snapshot {
	for {
		select {
		case <-ctx.Done():
			return c.snap
		case ll, ok := <-ch:
			if !ok {
				return c.snap
			}
			c.snap.Add(ll.Service, ll.Text)
		}
	}
}

// Snapshot returns the snapshot accumulated so far without stopping
// collection. Safe to call after Collect has returned.
func (c *Collector) Snapshot() *Snapshot {
	return c.snap
}
