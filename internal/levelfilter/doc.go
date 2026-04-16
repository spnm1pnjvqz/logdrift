// Package levelfilter provides severity-level-based filtering for log lines.
//
// It recognises four levels in ascending order: debug, info, warn, error.
// Level detection is keyword-based: the line text is scanned (case-insensitively)
// for the tokens ERROR, WARN, and INFO; lines matching none default to debug.
//
// Usage:
//
//	f, err := levelfilter.New("warn")   // pass warn and error only
//	if err != nil { ... }
//	out := f.Apply(ctx, in)
package levelfilter
