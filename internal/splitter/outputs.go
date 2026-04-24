package splitter

import "github.com/user/logdrift/internal/runner"

// MakeOutputs creates a buffered channel for every bucket name supplied.
// bufSize controls the channel buffer depth (0 for unbuffered).
func MakeOutputs(buckets []string, bufSize int) map[string]chan runner.LogLine {
	out := make(map[string]chan runner.LogLine, len(buckets))
	for _, b := range buckets {
		out[b] = make(chan runner.LogLine, bufSize)
	}
	return out
}

// BucketNames returns the sorted list of bucket names from an outputs map.
// Useful for deterministic iteration in tests.
func BucketNames(outputs map[string]chan runner.LogLine) []string {
	names := make([]string, 0, len(outputs))
	for k := range outputs {
		names = append(names, k)
	}
	return names
}
