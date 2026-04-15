// Package snapshot provides utilities for capturing, persisting, and loading
// point-in-time views of log streams produced by logdrift services.
//
// Typical usage:
//
//	// Capture
//	c := snapshot.NewCollector()
//	snap := c.Collect(ctx, fanInCh)
//	snap.Save("baseline.json")
//
//	// Restore
//	baseline, err := snapshot.Load("baseline.json")
//	if err != nil { ... }
package snapshot
