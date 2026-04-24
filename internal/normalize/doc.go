// Package normalize provides text normalization for log lines.
//
// A Normalizer can be configured to apply any combination of:
//
//   - Lowercase  – fold all characters to lower case
//   - CollapseSpaces – replace runs of whitespace with a single space
//   - Trim – strip leading and trailing whitespace
//
// Normalization is applied per-line and does not modify the Service
// field, making it safe to use in multi-service pipelines where
// service identity must be preserved.
package normalize
