// Package checkpoint persists the last-seen line offset for each service so
// that logdrift can resume tailing from where it left off after a restart.
//
// A Checkpoint is safe for concurrent use. Offsets are flushed to disk via
// Save and restored via Load.
package checkpoint

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
)

// Entry holds the persisted state for a single service.
type Entry struct {
	Service   string    `json:"service"`
	Offset    int64     `json:"offset"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Checkpoint tracks per-service byte offsets.
type Checkpoint struct {
	mu      sync.RWMutex
	entries map[string]*Entry
}

// New returns an empty Checkpoint.
func New() *Checkpoint {
	return &Checkpoint{
		entries: make(map[string]*Entry),
	}
}

// Set records offset for the given service, overwriting any previous value.
func (c *Checkpoint) Set(service string, offset int64) error {
	if service == "" {
		return errors.New("checkpoint: service name must not be empty")
	}
	if offset < 0 {
		return errors.New("checkpoint: offset must be non-negative")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[service] = &Entry{
		Service:   service,
		Offset:    offset,
		UpdatedAt: time.Now().UTC(),
	}
	return nil
}

// Get returns the stored offset for service, or 0 if none exists.
func (c *Checkpoint) Get(service string) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if e, ok := c.entries[service]; ok {
		return e.Offset
	}
	return 0
}

// Entries returns a snapshot of all stored entries.
func (c *Checkpoint) Entries() []Entry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Entry, 0, len(c.entries))
	for _, e := range c.entries {
		out = append(out, *e)
	}
	return out
}

// Save writes the checkpoint state to path as JSON, creating or truncating the
// file as needed.
func (c *Checkpoint) Save(path string) error {
	c.mu.RLock()
	entries := make([]Entry, 0, len(c.entries))
	for _, e := range c.entries {
		entries = append(entries, *e)
	}
	c.mu.RUnlock()

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Load reads a previously saved checkpoint file and returns a populated
// Checkpoint. If path does not exist an empty Checkpoint is returned without
// error, allowing callers to treat a missing file as a fresh start.
func Load(path string) (*Checkpoint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return New(), nil
		}
		return nil, err
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	c := New()
	for _, e := range entries {
		if e.Service == "" {
			continue
		}
		c.entries[e.Service] = &Entry{
			Service:   e.Service,
			Offset:    e.Offset,
			UpdatedAt: e.UpdatedAt,
		}
	}
	return c, nil
}
