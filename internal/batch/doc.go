// Package batch provides a stream processor that groups runner.LogLine values
// into fixed-size or time-bounded slices.
//
// A Batch is created with a maximum size and a flush interval.  Lines arriving
// on the input channel are accumulated in an internal buffer.  The buffer is
// flushed — emitted as a single []runner.LogLine — as soon as either condition
// is met:
//
//   - The buffer contains size lines, or
//   - interval has elapsed since the first line was added to the current buffer.
//
// Any lines remaining when the input channel is closed are flushed immediately
// before the output channel is closed.
package batch
