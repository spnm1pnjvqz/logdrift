// Package grep provides pattern-based filtering of log lines in real time.
//
// A Grep instance compiles one or more regular expressions and exposes an
// Apply method that reads from an input channel and forwards only lines
// whose text matches (or, in invert mode, does not match) any pattern.
//
// Usage:
//
//	g, err := grep.New([]string{"error", "fatal"}, false)
//	if err != nil { ... }
//	filtered := g.Apply(ctx, linesCh)
package grep
