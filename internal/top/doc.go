// Package top provides frequency tracking for log lines per service.
//
// It counts how often each unique line appears and exposes a Top(n)
// query so operators can identify the noisiest or most recurring
// messages in a live log stream.
//
// Usage:
//
//	tr, err := top.New(10)
//	out := tr.Apply(ctx, linesCh)
//	// later:
//	entries := tr.Top("api")
package top
