// Package differ provides log-line drift detection for logdrift.
//
// A Differ tracks lines that have been seen across services and classifies
// new lines as either "drift" (novel) or "non-drift" (already observed).
//
// Three modes are supported:
//
//	none  – drift detection is disabled; every line is non-drift.
//	uniq  – exact match after whitespace/case normalisation.
//	fuzzy – bigram Dice-coefficient similarity with a configurable threshold.
//
// A Pipeline wraps a Differ and connects it to a channel-based data flow,
// reading Lines from an input channel and emitting annotated Events.
package differ
