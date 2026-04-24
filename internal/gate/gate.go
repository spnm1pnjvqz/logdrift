// Package gate provides a conditional pass-through filter that opens or closes
// a log line stream based on whether a trigger pattern is matched.
//
// When the gate is closed, lines are dropped. When open, lines pass through.
// An open-pattern opens the gate; a close-pattern closes it again.
package gate

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/user/logdrift/internal/runner"
)

// Gate holds compiled trigger patterns and the current open/closed state.
type Gate struct {
	openRe  *regexp.Regexp
	closeRe *regexp.Regexp

	mu     sync.Mutex
	isOpen bool
}

// Config holds the configuration for a Gate.
type Config struct {
	// OpenPattern is a regular expression that, when matched, opens the gate.
	// Required.
	OpenPattern string

	// ClosePattern is a regular expression that, when matched, closes the gate.
	// Optional. If empty the gate stays open once triggered.
	ClosePattern string

	// InitiallyOpen controls whether lines are passed through before the first
	// open-pattern match. Defaults to false (gate starts closed).
	InitiallyOpen bool
}

// New creates a Gate from the given Config.
// Returns an error if OpenPattern is empty or any pattern fails to compile.
func New(cfg Config) (*Gate, error) {
	if cfg.OpenPattern == "" {
		return nil, fmt.Errorf("gate: OpenPattern must not be empty")
	}

	openRe, err := regexp.Compile(cfg.OpenPattern)
	if err != nil {
		return nil, fmt.Errorf("gate: invalid OpenPattern: %w", err)
	}

	var closeRe *regexp.Regexp
	if cfg.ClosePattern != "" {
		closeRe, err = regexp.Compile(cfg.ClosePattern)
		if err != nil {
			return nil, fmt.Errorf("gate: invalid ClosePattern: %w", err)
		}
	}

	return &Gate{
		openRe:  openRe,
		closeRe: closeRe,
		isOpen:  cfg.InitiallyOpen,
	}, nil
}

// Allow evaluates line against the gate's patterns, updates state, and returns
// whether the line should be forwarded downstream.
//
// State transitions happen before the allow decision so that the triggering
// line itself is included in (or excluded from) output.
func (g *Gate) Allow(line runner.LogLine) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Check open trigger first so the opening line passes through.
	if !g.isOpen && g.openRe.MatchString(line.Text) {
		g.isOpen = true
	}

	if !g.isOpen {
		return false
	}

	// Check close trigger; the closing line still passes through.
	if g.closeRe != nil && g.closeRe.MatchString(line.Text) {
		g.isOpen = false
	}

	return true
}

// Apply reads lines from src, forwards those that pass the gate, and closes
// the output channel when src is exhausted or ctx is cancelled.
func (g *Gate) Apply(ctx interface{ Done() <-chan struct{} }, src <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-src:
				if !ok {
					return
				}
				if g.Allow(line) {
					select {
					case out <- line:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out
}
