//go:build windows

package runner

import (
	"os"
	"os/exec"
)

func configureCommand(cmd *exec.Cmd) {}

func shutdownSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}

func interruptCommand(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return cmd.Process.Signal(os.Interrupt)
}

func killCommand(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return cmd.Process.Kill()
}

func processTreeStopped(proc *Process) bool {
	if proc == nil {
		return true
	}
	select {
	case <-proc.done:
		return true
	default:
		return false
	}
}
