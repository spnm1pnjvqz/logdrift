// Package snapshot captures and persists point-in-time log line state
// so that logdrift can compare current output against a baseline.
package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Entry holds a single captured log line with metadata.
type Entry struct {
	Service   string    `json:"service"`
	Line      string    `json:"line"`
	CapturedAt time.Time `json:"captured_at"`
}

// Snapshot is an ordered collection of log entries.
type Snapshot struct {
	CreatedAt time.Time `json:"created_at"`
	Entries   []Entry   `json:"entries"`
}

// New returns an empty Snapshot stamped with the current time.
func New() *Snapshot {
	return &Snapshot{CreatedAt: time.Now()}
}

// Add appends a new entry to the snapshot.
func (s *Snapshot) Add(service, line string) {
	s.Entries = append(s.Entries, Entry{
		Service:    service,
		Line:       line,
		CapturedAt: time.Now(),
	})
}

// Save writes the snapshot as JSON to the given file path.
func (s *Snapshot) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("snapshot: create %q: %w", path, err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

// Load reads a snapshot from the given file path.
func Load(path string) (*Snapshot, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("snapshot: open %q: %w", path, err)
	}
	defer f.Close()
	var s Snapshot
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return nil, fmt.Errorf("snapshot: decode %q: %w", path, err)
	}
	return &s, nil
}
