// Package splitter provides pattern-based routing of log lines into named
// output channels.
//
// Usage:
//
//	s, err := splitter.New(map[string]string{
//		"errors": `(?i)error|fatal`,
//		"slow":   `latency_ms:[0-9]{4,}`,
//	}, "other")
//
//	outputs := splitter.MakeOutputs([]string{"errors", "slow", "other"}, 64)
//	s.Apply(ctx, src, outputs)
//
// Lines that match no rule are forwarded to the defaultBucket channel (if
// present in outputs). Lines whose resolved bucket is not in outputs are
// silently dropped, which allows callers to discard uninteresting traffic by
// simply omitting that bucket from the outputs map.
package splitter
