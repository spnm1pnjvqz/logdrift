// Package fork provides a fan-out primitive for log-line channels.
//
// Fork takes a single source channel and replicates every line to N
// independent output channels, allowing separate pipeline branches to
// consume the same stream without interference.
package fork
