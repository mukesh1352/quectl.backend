package queue

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"time"
)

// ExecResult stores detailed information about a command execution.
type ExecResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Err      error
	Duration time.Duration // ✅ Added field for job duration tracking
}

// ExecCommand runs a shell command with a timeout and returns its results.
func ExecCommand(command string, timeout time.Duration) ExecResult {
	start := time.Now() // ✅ Record when command starts

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start) // ✅ Measure total execution time

	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			err = errors.New("job timeout exceeded")
			exitCode = -1
		} else {
			exitCode = 1
		}
	}

	return ExecResult{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Err:      err,
		Duration: duration,
	}
}
