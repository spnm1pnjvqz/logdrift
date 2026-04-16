// Package colorize provides per-service ANSI color assignment for log lines.
//
// A Colorizer assigns a color from a fixed palette to each unique service name
// encountered. Colors are assigned in order of first appearance and remain
// stable for the lifetime of the Colorizer.
//
// Apply wraps the Text field of every runner.LogLine passing through a channel
// with the service's assigned color, making interleaved output easier to scan.
package colorize
