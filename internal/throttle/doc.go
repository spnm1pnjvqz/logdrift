// Package throttle implements a simple token-bucket throttle for log line
// channels. It limits the number of lines forwarded per second, which is
// useful when a noisy service would otherwise flood the display or the
// differ pipeline.
//
// Usage:
//
//	th, err := throttle.New(50) // at most 50 lines/sec
//	if err != nil { ... }
//	throttled := th.Apply(ctx, rawLines)
package throttle
