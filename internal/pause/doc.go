// Package pause implements a pausable pipeline stage for log line streams.
//
// A Controller can be shared across goroutines to pause and resume the
// forwarding of runner.LogLine values without dropping any lines. Lines
// received while paused are buffered in the input channel (up to its
// capacity) and delivered in order once the controller is resumed.
//
// Typical usage:
//
//	ctrl := pause.New()
//	out := pause.Apply(ctx, ctrl, in)
//	// later, from another goroutine:
//	ctrl.Pause()
//	ctrl.Resume()
package pause
