package label

import (
	"context"
	"fmt"

	"github.com/user/logdrift/internal/runner"
)

// ServiceChannel pairs a service name with its log line source.
type ServiceChannel struct {
	Service string
	Ch      <-chan runner.LogLine
}

// LabelAll attaches the appropriate service label to each ServiceChannel and
// returns a slice of labelled output channels ready for fan-in.
func LabelAll(ctx context.Context, sources []ServiceChannel) ([]<-chan runner.LogLine, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("label: no sources provided")
	}
	out := make([]<-chan runner.LogLine, 0, len(sources))
	for _, src := range sources {
		lbr, err := New(src.Service)
		if err != nil {
			return nil, fmt.Errorf("label: %w", err)
		}
		out = append(out, lbr.Apply(ctx, src.Ch))
	}
	return out, nil
}
