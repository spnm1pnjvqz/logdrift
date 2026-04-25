package batch

import (
	"context"

	"github.com/user/logdrift/internal/runner"
)

// ApplyContext is a convenience wrapper around Batch.Apply that accepts a
// concrete *context.Context instead of the interface used internally, making
// call-sites that already hold a context.Context slightly more ergonomic.
func ApplyContext(ctx context.Context, b *Batch, src <-chan runner.LogLine) <-chan []runner.LogLine {
	return b.Apply(ctx, src)
}
