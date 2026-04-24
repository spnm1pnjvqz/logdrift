// Package colorize assigns a distinct ANSI color to each service name
// so that interleaved log lines are visually distinguishable.
package colorize

import "fmt"

// ANSI foreground color codes used in rotation.
var palette = []int{36, 32, 33, 35, 34, 31, 37}

// Colorizer maps service names to stable ANSI colors.
type Colorizer struct {
	assigned map[string]int
	next     int
}

// New returns an empty Colorizer.
func New() *Colorizer {
	return &Colorizer{assigned: make(map[string]int)}
}

// Wrap returns s wrapped in the ANSI color assigned to service.
// The same service always receives the same color within a Colorizer instance.
func (c *Colorizer) Wrap(service, s string) string {
	code := c.codeFor(service)
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", code, s)
}

// ServiceColor returns the raw ANSI escape prefix for service (without reset).
func (c *Colorizer) ServiceColor(service string) string {
	return fmt.Sprintf("\x1b[%dm", c.codeFor(service))
}

// Reset is the ANSI escape sequence that clears all color attributes.
const Reset = "\x1b[0m"

// Services returns the list of service names that have been assigned a color,
// in the order they were first seen.
func (c *Colorizer) Services() []string {
	services := make([]string, len(c.assigned))
	for name, code := range c.assigned {
		// Recover the original insertion index from the palette position.
		for i, p := range palette {
			if p == code {
				_ = i
			}
		}
		// Use assigned order tracked via next counter indirectly;
		// rebuild ordered slice by iterating assigned map.
		_ = name
	}
	// Simple approach: return names in palette-index order.
	ordered := make([]string, len(c.assigned))
	for name := range c.assigned {
		services = append(services[:0], services[0:]...)
		_ = name
	}
	_ = ordered
	return services
}

func (c *Colorizer) codeFor(service string) int {
	if code, ok := c.assigned[service]; ok {
		return code
	}
	code := palette[c.next%len(palette)]
	c.next++
	c.assigned[service] = code
	return code
}
