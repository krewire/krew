package commands

import (
	"log/slog"
	"os"
	"os/exec"
	"syscall"

	"github.com/krewire/krewire/pkg/env"
)

var currentChild *exec.Cmd

// restartApp starts the app binary as a child, streaming its output and
// exporting KIW_ENV/KIW_DEBUG alongside APP_ADDR (KWL-K4T7W).
func restartApp(bin, root, addr string, env env.Env, debug bool) {
	cmd := exec.Command(bin)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = childEnviron(os.Environ(), addr, env, debug)
	if err := cmd.Start(); err != nil {
		slog.Error("failed to start app", "error", err)
		return
	}
	currentChild = cmd
	slog.Info("app started", "pid", cmd.Process.Pid)
	go func() {
		if err := cmd.Wait(); err != nil {
			slog.Warn("app exited", "pid", cmd.Process.Pid, "error", err)
		}
	}()
}

// stopChild sends SIGINT to the running child and waits for it to exit.
func stopChild() {
	cmd := currentChild
	if cmd == nil || cmd.Process == nil {
		return
	}
	currentChild = nil
	slog.Info("stopping app", "pid", cmd.Process.Pid)
	_ = cmd.Process.Signal(syscall.SIGINT)
	_ = cmd.Wait()
}
