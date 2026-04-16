// Package alert provides pattern-based alerting for log streams.
//
// An Alerter is initialised with a map of named regular expressions.
// Each LogLine passing through Apply is tested against every rule;
// matching lines produce an Event carrying the rule name, service, and
// original line text.
//
// Typical usage:
//
//	a, err := alert.New(map[string]string{
//		"error": `(?i)error`,
//		"oom":   `out of memory`,
//	})
//	events := alert.Apply(ctx, a, logLines)
package alert
