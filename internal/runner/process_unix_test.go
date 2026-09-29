//go:build !windows

package runner

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestShutdownSignalsIncludeHangup(t *testing.T) {
	signals := shutdownSignals()

	for _, want := range []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGHUP} {
		found := false
		for _, got := range signals {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected shutdown signals to include %v, got %v", want, signals)
		}
	}
}

func TestStopKillsProcessGroupAfterShellExits(t *testing.T) {
	pidFile := t.TempDir() + "/child.pid"
	command := "trap 'exit 0' INT; " +
		"sh -c 'trap \"\" INT TERM; while :; do sleep 1; done' & " +
		"child=$!; echo $child > " + strconv.Quote(pidFile) + "; wait"

	proc, err := start(t.TempDir(), command)
	if err != nil {
		t.Fatalf("start returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = killCommand(proc.cmd)
	})

	childPID := waitForPIDFile(t, pidFile)
	if err := syscall.Kill(childPID, 0); err != nil {
		t.Fatalf("expected child process %d to be running: %v", childPID, err)
	}

	proc.Stop()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		err := syscall.Kill(childPID, 0)
		if errors.Is(err, syscall.ESRCH) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("child process %d survived Process.Stop", childPID)
}

func waitForPIDFile(t *testing.T, path string) int {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil {
			pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
			if err != nil {
				t.Fatalf("parse child PID: %v", err)
			}
			return pid
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("read child PID: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for child PID file %s", path)
	return 0
}
