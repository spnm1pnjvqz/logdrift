// Package columns provides a log line formatter that splits each line on a
// configurable delimiter and pads the resulting fields to fixed widths,
// producing neatly aligned tabular output in the terminal.
//
// Usage:
//
//	f, err := columns.New(" ", []int{10, 20, 40})
//	if err != nil { ... }
//	out := f.Apply(src)
package columns
