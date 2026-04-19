// Package linenum provides a Stamper that prepends a per-service
// incrementing line number to every LogLine passing through the pipeline.
//
// Usage:
//
//	s := linenum.New(4)          // 4-digit zero-padded counter
//	out := s.Apply(ctx, src)
//
// Each service maintains its own independent counter, so line numbers
// restart at 1 when a new service name is first seen.
package linenum
