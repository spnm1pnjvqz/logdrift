// Package lineformat provides template-based formatting for log lines.
//
// A Formatter is created with a template string containing one or more
// of the following placeholders:
//
//   - {service}  the originating service name
//   - {text}     the raw log line text
//   - {time}     the current UTC time in RFC 3339 format
//
// Example template: "[{service}] {time} — {text}"
//
// Apply wraps a Formatter in a streaming pipeline stage that transforms
// every incoming LogLine and forwards it on the output channel.
package lineformat
