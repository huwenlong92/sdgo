//go:build windows

package runner

import (
	"os"
	"os/exec"
)

func configureCommand(cmd *exec.Cmd) {}

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
