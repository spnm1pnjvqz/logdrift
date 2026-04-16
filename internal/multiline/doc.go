// Package multiline provides a Joiner that coalesces multi-line log
// records (e.g. Java stack traces, Python tracebacks) into a single
// runner.LogLine before the line is forwarded to the rest of the
// logdrift pipeline.
//
// A logical record begins when an incoming line matches the configured
// start regular expression. Subsequent lines that do NOT match the
// start pattern are treated as continuations and are appended to the
// current record, separated by newlines.
//
// A configurable timeout ensures that a pending record is flushed even
// when no new start line arrives within the expected window, preventing
// the pipeline from stalling on the last record in a quiet stream.
package multiline
