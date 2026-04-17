// Package linecount tracks the number of log lines emitted per service.
//
// Usage:
//
//	lc := linecount.New()
//	lc.Add("api", line)
//	count := lc.Get("api")
//	snapshot := lc.Snapshot()
package linecount
