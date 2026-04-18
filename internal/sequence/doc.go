// Package sequence provides a Sequencer that appends a monotonically
// increasing counter tag to every log line passing through a stream.
//
// This makes it easy to spot gaps when multiple services are merged:
//
//	seq, _ := sequence.New("#")
//	out := seq.Apply(ctx, in)
//
// Each emitted line will have a suffix like " [#42]" appended to its Text
// field. The counter is global across all services handled by a single
// Sequencer instance, so ordering across merged channels is observable.
package sequence
