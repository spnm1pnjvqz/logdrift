// Package levelfilter drops or passes log lines based on a minimum severity level.
package levelfilter

import (
	"fmt"
	"strings"

	"github.com/yourorg/logdrift/internal/runner"
)

// Level represents a log severity level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var levelNames = map[string]Level{
	"debug": LevelDebug,
	"info":  LevelInfo,
	"warn":  LevelWarn,
	"error": LevelError,
}

// Filter passes only lines whose detected level is >= the minimum level.
type Filter struct {
	min Level
}

// New returns a Filter with the given minimum level string (case-insensitive).
func New(minLevel string) (*Filter, error) {
	l, ok := levelNames[strings.ToLower(minLevel)]
	if !ok {
		return nil, fmt.Errorf("levelfilter: unknown level %q", minLevel)
	}
	return &Filter{min: l}, nil
}

// detect returns the severity level found in the line text, defaulting to Debug.
func (f *Filter) detect(text string) Level {
	upper := strings.ToUpper(text)
	switch {
	case strings.Contains(upper, "ERROR"):
		return LevelError
	case strings.Contains(upper, "WARN"):
		return LevelWarn
	case strings.Contains(upper, "INFO"):
		return LevelInfo
	default:
		return LevelDebug
	}
}

// Allow reports whether the line meets the minimum level threshold.
func (f *Filter) Allow(text string) bool {
	return f.detect(text) >= f.min
}

// Apply reads from in and forwards lines that pass the level filter to the returned channel.
func (f *Filter) Apply(ctx interface{ Done() <-chan struct{} }, in <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-in:
				if !ok {
					return
				}
				if f.Allow(line.Text) {
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
