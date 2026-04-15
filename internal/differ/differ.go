package differ

import (
	"strings"
	"time"
)

// DiffMode controls how lines are compared across services.
type DiffMode string

const (
	DiffModeNone   DiffMode = "none"
	DiffModeUniq   DiffMode = "uniq"
	DiffModeFuzzy  DiffMode = "fuzzy"
)

// Line represents a log line from a named service.
type Line struct {
	Service   string
	Text      string
	Timestamp time.Time
}

// Differ compares log lines across services and emits those that diverge.
type Differ struct {
	mode    DiffMode
	seen    map[string]struct{}
	thresh  float64
}

// New creates a Differ with the given mode. thresh is used for fuzzy mode (0–1).
func New(mode DiffMode, thresh float64) *Differ {
	if thresh <= 0 {
		thresh = 0.8
	}
	return &Differ{
		mode:   mode,
		seen:   make(map[string]struct{}),
		thresh: thresh,
	}
}

// IsDrift returns true when the line is considered a drift event
// relative to lines already seen from other services.
func (d *Differ) IsDrift(l Line) bool {
	switch d.mode {
	case DiffModeNone:
		return false
	case DiffModeFuzzy:
		return d.fuzzyDrift(l.Text)
	default: // uniq
		return d.uniqDrift(l.Text)
	}
}

// Record stores the normalised line so future calls to IsDrift can compare.
func (d *Differ) Record(l Line) {
	d.seen[normalise(l.Text)] = struct{}{}
}

func (d *Differ) uniqDrift(text string) bool {
	_, ok := d.seen[normalise(text)]
	return !ok
}

func (d *Differ) fuzzyDrift(text string) bool {
	norm := normalise(text)
	for k := range d.seen {
		if similarity(norm, k) >= d.thresh {
			return false
		}
	}
	return true
}

func normalise(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

// similarity returns a simple bigram-based Dice coefficient.
func similarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	if len(a) < 2 || len(b) < 2 {
		return 0.0
	}
	bA := bigrams(a)
	bB := bigrams(b)
	var shared int
	for k, n := range bA {
		if m, ok := bB[k]; ok {
			if n < m {
				shared += n
			} else {
				shared += m
			}
		}
	}
	return float64(2*shared) / float64(len(bA)+len(bB))
}

func bigrams(s string) map[string]int {
	m := make(map[string]int)
	for i := 0; i < len(s)-1; i++ {
		m[s[i:i+2]]++
	}
	return m
}
