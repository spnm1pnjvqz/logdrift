package differ

import "context"

// Event is emitted by the Pipeline for every line that passes the drift filter.
type Event struct {
	Line  Line
	Drift bool
}

// Pipeline reads from a merged channel of Lines, runs them through a Differ,
// and emits Events on the output channel.
type Pipeline struct {
	differ *Differ
}

// NewPipeline constructs a Pipeline backed by the given Differ.
func NewPipeline(d *Differ) *Pipeline {
	return &Pipeline{differ: d}
}

// Run consumes lines from in and sends Events to the returned channel.
// The output channel is closed when ctx is done or in is closed.
func (p *Pipeline) Run(ctx context.Context, in <-chan Line) <-chan Event {
	out := make(chan Event, 64)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case l, ok := <-in:
				if !ok {
					return
				}
				drift := p.differ.IsDrift(l)
				p.differ.Record(l)
				select {
				case out <- Event{Line: l, Drift: drift}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
