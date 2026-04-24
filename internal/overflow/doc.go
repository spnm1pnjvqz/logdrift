// Package overflow implements a back-pressure limiter for log-line channels.
//
// Two policies are supported:
//
//   - Drop: when the internal buffer is full, incoming lines are silently
//     discarded so that the producer is never blocked.
//
//   - Block: the Apply goroutine waits until the consumer catches up, providing
//     lossless delivery at the cost of potential producer stalls.
//
// Usage:
//
//	lim, err := overflow.New(512, overflow.Drop)
//	if err != nil { ... }
//	filtered := lim.Apply(ctx, src)
package overflow
