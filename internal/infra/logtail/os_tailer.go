package logtail

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
)

type OSTailer struct{}

func (OSTailer) Tail(ctx context.Context, path string, lines int) (io.ReadCloser, error) {
	if path == "" {
		return nil, fmt.Errorf("log path is required")
	}
	if lines <= 0 {
		lines = 100
	}

	cmd := exec.CommandContext(ctx, "tail", "-f", "-n", strconv.Itoa(lines), path)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		_ = stdout.Close()
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		_ = stdout.Close()
		_ = stderr.Close()
		return nil, err
	}

	// Return a read-closer that cancels the command via context and waits.
	return &cmdReadCloser{
		stdout: stdout,
		stderr: stderr,
		wait:   cmd.Wait,
	}, nil
}

type cmdReadCloser struct {
	stdout io.ReadCloser
	stderr io.ReadCloser
	wait   func() error
}

func (c *cmdReadCloser) Read(p []byte) (int, error) {
	return c.stdout.Read(p)
}

func (c *cmdReadCloser) Close() error {
	_ = c.stdout.Close()
	_ = c.stderr.Close()
	_ = c.wait()
	return nil
}
