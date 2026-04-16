// Package window implements a per-service sliding time-window counter.
//
// Use New to create a Window with a fixed rolling duration, then call Add
// each time a log line is observed and Count to query how many lines have
// arrived within that window for a given service.
//
// All methods are safe for concurrent use.
package window
