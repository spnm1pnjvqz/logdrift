// Package retry wraps a log-line source factory with a retry policy.
//
// When the source channel closes unexpectedly (i.e. while the context is still
// active), Apply will re-invoke the factory up to MaxAttempts times, waiting
// Delay between each attempt. This is useful for commands that crash and need
// to be restarted, or file tails that temporarily disappear during a rotation.
//
// Usage:
//
//	r, err := retry.New(retry.Config{MaxAttempts: 3, Delay: time.Second})
//	ch, err := r.Apply(ctx, myFactory)
package retry
