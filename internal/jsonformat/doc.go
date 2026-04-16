// Package jsonformat provides a pipeline stage that detects JSON log lines and
// pretty-prints them for human-readable display.
//
// Non-JSON lines pass through unchanged, so the formatter is safe to insert
// into any pipeline regardless of whether the upstream service emits structured
// logs.
//
// Usage:
//
//	f := jsonformat.New("  ") // two-space indent
//	out := f.Apply(ctx, in)
package jsonformat
