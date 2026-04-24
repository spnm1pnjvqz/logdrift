// Package sample implements periodic log-line sampling.
//
// It forwards one out of every N lines received per service, allowing
// high-volume streams to be thinned before downstream processing.
//
// Usage:
//
//	s, err := sample.New(10) // keep every 10th line
//	if err != nil {
//		log.Fatal(err)
//	}
//	filtered := s.Apply(ctx, src)
package sample
