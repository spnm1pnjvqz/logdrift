// Package redact provides regex-based redaction of sensitive values in log
// lines before they are displayed or stored.
//
// Usage:
//
//	r, err := redact.New(map[string]string{
//		`password=\S+`: "password=[REDACTED]",
//		`token=\S+`:    "token=[REDACTED]",
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	clean := r.Apply(rawLine)
//
// ApplyToChannel wraps a string channel so that every line is redacted
// transparently as it flows through the pipeline.
package redact
