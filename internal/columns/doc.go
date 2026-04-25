// Package columns provides a log line formatter that splits each line on a
// configurable delimiter and pads the resulting fields to fixed widths,
// producing neatly aligned tabular output in the terminal.
//
// Each field is left-aligned and truncated or padded with spaces to match the
// corresponding width. If a line contains more fields than configured widths,
// the extra fields are appended without padding. If a line contains fewer
// fields, the missing columns are rendered as empty padded strings.
//
// Usage:
//
//	f, err := columns.New(" ", []int{10, 20, 40})
//	if err != nil { ... }
//	out := f.Apply(src)
//
// Errors:
//
// New returns an error if the delimiter is empty or if any width value is
// less than or equal to zero.
package columns
