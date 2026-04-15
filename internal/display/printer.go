// Package display handles terminal output formatting for logdrift.
package display

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"

	"github.com/yourorg/logdrift/internal/differ"
)

// Palette maps service names to a fixed set of distinct colors.
var palette = []*color.Color{
	color.New(color.FgCyan),
	color.New(color.FgGreen),
	color.New(color.FgMagenta),
	color.New(color.FgYellow),
	color.New(color.FgBlue),
	color.New(color.FgHiCyan),
	color.New(color.FgHiGreen),
}

// Printer writes formatted log events to an output writer.
type Printer struct {
	out      io.Writer
	mu       sync.Mutex
	colorMap map[string]*color.Color
	next     int
	driftFn  *color.Color
}

// New creates a Printer that writes to out.
func New(out io.Writer) *Printer {
	if out == nil {
		out = os.Stdout
	}
	return &Printer{
		out:     out,
		colorMap: make(map[string]*color.Color),
		driftFn:  color.New(color.FgHiRed, color.Bold),
	}
}

// colorFor returns a stable color for the given service name.
func (p *Printer) colorFor(service string) *color.Color {
	if c, ok := p.colorMap[service]; ok {
		return c
	}
	c := palette[p.next%len(palette)]
	p.next++
	p.colorMap[service] = c
	return c
}

// Print formats and writes a single differ.Event to the output.
func (p *Printer) Print(ev differ.Event) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ts := ev.Line.Timestamp.Format(time.RFC3339)
	svcColor := p.colorFor(ev.Line.Service)
	svcLabel := svcColor.Sprintf("%-15s", ev.Line.Service)

	var driftTag string
	if ev.IsDrift {
		driftTag = p.driftFn.Sprint(" [DRIFT]")
	}

	fmt.Fprintf(p.out, "%s %s %s%s\n",
		ts,
		svcLabel,
		strings.TrimRight(ev.Line.Text, "\r\n"),
		driftTag,
	)
}

// Run consumes events from ch until it is closed or ctx is done.
func (p *Printer) Run(events <-chan differ.Event) {
	for ev := range events {
		p.Print(ev)
	}
}
