// Package tee provides a Tee function that duplicates a LogLine channel into
// two independent output channels. Both outputs receive every line emitted by
// the source, enabling fan-out processing pipelines without channel re-use.
//
// Example:
//
//	a, b := tee.Tee(ctx, src)
//	go pipeline1(a) // e.g. write to file
//	go pipeline2(b) // e.g. forward to differ
package tee
