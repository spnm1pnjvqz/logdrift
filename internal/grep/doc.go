// Package grep provides pattern-based filtering of log lines in real time.
//
// A Grep instance compiles one or more regular expressions and exposes an
// Apply method that reads from an input channel and forwards only lines
// whose text matches (or, in invert mode, does not match) any pattern.
//
// Invert mode (analogous to grep -v) causes Apply to forward only lines
// that do NOT match any of the provided patterns.
//
// Concurrency: Apply is safe to call from a single goroutine. The returned
// channel is closed automatically when the input channel is closed or the
// provided context is cancelled, whichever occurs first.
//
// Usage:
//
//	g, err := grep.New([]string{"error", "fatal"}, false)
//	if err != nil { ... }
//	filtered := g.Apply(ctx, linesCh)
//
// Invert example:
//
//	g, err := grep.New([]string{"debug"}, true)
//	if err != nil { ... }
//	filtered := g.Apply(ctx, linesCh) // drops lines containing "debug"
//
// Pattern syntax follows the RE2 regular expression syntax as documented at
// https://golang.org/s/re2syntax. Patterns are compiled once at construction
// time; any compilation error is returned by New before processing begins.
package grep
