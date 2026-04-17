// Package pause provides a pausable channel pipeline stage that can
// temporarily halt the flow of log lines without dropping them.
package pause

import (
	"sync"

	"github.com/user/logdrift/internal/runner"
)

// Controller allows callers to pause and resume a pipeline stage.
type Controller struct {
	mu     sync.Mutex
	paused bool
	cond   *sync.Cond
}

// New returns a new Controller in the resumed state.
func New() *Controller {
	c := &Controller{}
	c.cond = sync.NewCond(&c.mu)
	return c
}

// Pause halts line forwarding until Resume is called.
func (c *Controller) Pause() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.paused = true
}

// Resume restores line forwarding.
func (c *Controller) Resume() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.paused = false
	c.cond.Broadcast()
}

// IsPaused reports whether the controller is currently paused.
func (c *Controller) IsPaused() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.paused
}

// Apply reads lines from in, blocks while paused, and forwards to the
// returned channel. The output channel is closed when ctx is done or in
// is closed.
func Apply(ctx interface{ Done() <-chan struct{} }, ctrl *Controller, in <-chan runner.LogLine) <-chan runner.LogLine {
	out := make(chan runner.LogLine)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-in:
				if !ok {
					return
				}
				ctrl.mu.Lock()
				for ctrl.paused {
					ctrl.cond.Wait()
				}
				ctrl.mu.Unlock()
				select {
				case out <- line:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
