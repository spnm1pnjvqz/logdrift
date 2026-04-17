// Package ratelimit applies a token-bucket rate limit to a stream of log lines.
//
// A zero rate means unlimited. Positive rates are expressed as lines per second.
//
// Usage:
//
//	rl, err := ratelimit.New(100)
//	if err != nil {
//		log.Fatal(err)
//	}
//	out := rl.Apply(ctx, in)
package ratelimit
