// Package filter provides line-level include/exclude filtering
// for log streams using regular expression patterns.
//
// A Filter is constructed with optional include and exclude pattern slices.
// When include patterns are provided, only lines matching at least one
// pattern are passed through. Exclude patterns are applied after include
// patterns and remove any matching lines.
//
// Apply wires a Filter into a pipeline by reading from an input channel
// and writing accepted lines to the returned output channel.
package filter
