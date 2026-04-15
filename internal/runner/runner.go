package runner

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"sync"
)

// Line represents a single log line from a named service.
type Line struct {
	Service string
	Text    string
	Err     error
}

// Runner starts service commands and streams their output as Lines.
type Runner struct {
	mu   sync.Mutex
	cmds []*exec.Cmd
}

// New creates a new Runner.
func New() *Runner {
	return &Runner{}
}

// Start launches a command for the given service and streams stdout/stderr
// lines into the returned channel. The channel is closed when the command exits
// or the context is cancelled.
func (r *Runner) Start(ctx context.Context, service, shell string, args []string) (<-chan Line, error) {
	cmd := exec.CommandContext(ctx, shell, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.cmds = append(r.cmds, cmd)
	r.mu.Unlock()

	ch := make(chan Line, 64)

	var wg sync.WaitGroup
	for _, pipe := range []io.Reader{stdout, stderr} {
		wg.Add(1)
		go func(rd io.Reader) {
			defer wg.Done()
			scanner := bufio.NewScanner(rd)
			for scanner.Scan() {
				ch <- Line{Service: service, Text: scanner.Text()}
			}
			if err := scanner.Err(); err != nil {
				ch <- Line{Service: service, Err: err}
			}
		}(pipe)
	}

	go func() {
		wg.Wait()
		_ = cmd.Wait()
		close(ch)
	}()

	return ch, nil
}

// StopAll terminates all running commands.
func (r *Runner) StopAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, cmd := range r.cmds {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}
}
