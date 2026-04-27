// Package debounce provides a Debouncer that suppresses rapid bursts of
// identical log lines from the same service.
//
// When the same text is emitted by a service multiple times within the
// configured quiet window, only the first occurrence is forwarded downstream.
// Each new occurrence resets the window timer. Once the window expires without
// a repeat, the next occurrence of that text is forwarded again.
//
// Usage:
//
//	d, err := debounce.New(500 * time.Millisecond)
//	out := debounce.Apply(ctx, d, src)
package debounce
