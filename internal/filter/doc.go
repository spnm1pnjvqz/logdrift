// Package filter provides real-time filtering of log lines produced by
// runner.Runner instances.
//
// A Filter is built from a Config that specifies include and exclude regular
// expression patterns:
//
//   - Include patterns: only lines matching at least one pattern are forwarded.
//     If no include patterns are configured, all lines pass this check.
//   - Exclude patterns: lines matching any exclude pattern are always dropped,
//     even if they also match an include pattern.
//
// Use Apply to wire a Filter into a pipeline between a fan-in channel and the
// differ pipeline.
package filter
