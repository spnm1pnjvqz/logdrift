// Package since provides a time-based filter for log lines.
//
// It parses the timestamp at the beginning of each log line and discards
// lines that predate a configured cutoff. Lines whose timestamps cannot
// be parsed are forwarded unchanged so no data is silently lost.
//
// Usage:
//
//	f, err := since.New(cutoff, nil)  // nil uses built-in format list
//	out := f.Apply(ctx, in)
package since
