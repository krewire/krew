package sdlc

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

// CommandExecutor executes a single shell command in the target directory.
type CommandExecutor func(ctx context.Context, dir, command string) (exitCode int, output string, err error)

// DefaultCommandExecutor executes commands via /bin/sh -c with bounded buffer output.
func DefaultCommandExecutor(ctx context.Context, dir, command string) (int, string, error) {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	cmd.Dir = dir

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	output := buf.String()
	// Bound output to max 64KB for verification evidence capture
	if len(output) > 64*1024 {
		output = output[:64*1024] + "\n...[truncated]..."
	}

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return exitCode, output, err
}

// Verifier executes verification commands and collects structured qualification evidence.
type Verifier struct {
	Dir      string
	Executor CommandExecutor
}

// NewVerifier creates a Verifier for the given directory.
func NewVerifier(dir string, execFn CommandExecutor) *Verifier {
	if execFn == nil {
		execFn = DefaultCommandExecutor
	}
	return &Verifier{
		Dir:      dir,
		Executor: execFn,
	}
}

// Run executes all given verification commands in order.
// If any command fails, allPassed will be false.
func (v *Verifier) Run(ctx context.Context, commands []string) ([]VerificationResult, bool) {
	var results []VerificationResult
	allPassed := true

	if len(commands) == 0 {
		return results, true
	}

	for _, cmdStr := range commands {
		cmdStr = strings.TrimSpace(cmdStr)
		if cmdStr == "" {
			continue
		}

		start := time.Now()
		exitCode, out, _ := v.Executor(ctx, v.Dir, cmdStr)
		duration := time.Since(start)

		passed := exitCode == 0
		if !passed {
			allPassed = false
		}

		results = append(results, VerificationResult{
			Command:   cmdStr,
			ExitCode:  exitCode,
			Passed:    passed,
			Output:    strings.TrimSpace(out),
			Duration:  duration,
			Timestamp: start,
		})
	}

	return results, allPassed
}
