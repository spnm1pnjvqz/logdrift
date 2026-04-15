// Package timestamp provides a Stamper that prepends formatted timestamps
// to log lines as they travel through the logdrift pipeline.
//
// Supported formats:
//
//   - rfc3339  — full RFC 3339 timestamp (e.g. 2024-01-15T10:04:05Z)
//   - unix     — seconds since the Unix epoch (e.g. 1705312245)
//   - kitchen  — 12-hour clock (e.g. 10:04AM)
//   - relative — elapsed time since the Stamper was created (e.g. +1.234s)
//
// Usage:
//
//	s, err := timestamp.New(timestamp.FormatRFC3339)
//	stampedCh := s.Apply(lineCh)
package timestamp
