// Package aggregate provides a sliding-window line counter for logdrift.
//
// Usage:
//
//	agg, err := aggregate.New(5 * time.Second)
//	if err != nil { ... }
//
//	summaries := agg.Apply(lineCh)
//	for s := range summaries {
//		fmt.Printf("%s: %d lines in window ending %s\n",
//			s.Key, s.Count, s.WindowEnd.Format(time.RFC3339))
//	}
//
// The aggregator tallies log lines by service name and emits a Summary for
// each service at the end of every window (or when the input channel closes).
package aggregate
