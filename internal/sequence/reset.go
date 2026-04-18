package sequence

import "sync/atomic"

// Reset sets the internal counter back to zero.
// Useful in tests or when a new session begins.
func (s *Sequencer) Reset() {
	atomic.StoreUint64(&s.counter, 0)
}

// Current returns the last sequence number that was issued.
// Returns 0 if no lines have been stamped yet.
func (s *Sequencer) Current() uint64 {
	return atomic.LoadUint64(&s.counter)
}
