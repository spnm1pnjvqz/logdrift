// Package mask provides pattern-based text masking for log lines.
//
// Create a Masker with one or more regular expressions and an optional
// placeholder string. Each pattern match in a log line's text is replaced
// with the placeholder, defaulting to "[MASKED]".
//
// Example:
//
//	m, err := mask.New([]string{`password=\S+`, `token=\S+`}, "")
//	masked := m.Apply(line.Text)
//
// Use Transform to apply masking inline in a pipeline channel.
package mask
