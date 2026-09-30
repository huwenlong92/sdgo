//go:build !windows

package runner

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

const runnerHelperEnv = "SDGO_RUNNER_HELPER"

func TestRunStopsCurrentProcessAfterRestart(t *testing.T) {
	projectDir := filepath.Join(t.TempDir(), "project")
	if err := os.Mkdir(projectDir, 0o755); err != nil {
		t.Fatalf("create project directory: %v", err)
	}
	writeRunnerTestFile(t, filepath.Join(projectDir, "go.mod"), "module example.com/runner-test\n")
	watchPath := filepath.Join(projectDir, "main.go")
	writeRunnerTestFile(t, watchPath, "package main\n")

	pidFile := filepath.Join(t.TempDir(), "service.pids")
	command := fmt.Sprintf(
		"echo $$ >> %s; trap 'exit 0' INT TERM HUP; while :; do sleep 1; done",
		strconv.Quote(pidFile),
	)

	helper := exec.Command(os.Args[0], "-test.run=^TestRunHelperProcess$")
	helper.Env = append(os.Environ(),
		runnerHelperEnv+"=1",
		"SDGO_RUNNER_HELPER_DIR="+projectDir,
		"SDGO_RUNNER_HELPER_COMMAND="+command,
	)
	helper.Stderr = os.Stderr
	if err := helper.Start(); err != nil {
		t.Fatalf("start runner helper: %v", err)
	}

	var servicePIDs []int
	t.Cleanup(func() {
		if helper.Process != nil {
			_ = helper.Process.Kill()
		}
		for _, pid := range servicePIDs {
			_ = syscall.Kill(-pid, syscall.SIGKILL)
		}
	})

	servicePIDs = waitForServicePIDs(t, pidFile, 1)
	writeRunnerTestFile(t, watchPath, "package main\n\n// restart\n")
	servicePIDs = waitForServicePIDs(t, pidFile, 2)

	if err := helper.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("interrupt runner helper: %v", err)
	}
	waitForHelperExit(t, helper)

	for _, pid := range servicePIDs {
		waitForProcessExit(t, pid)
	}
}

func TestRunHelperProcess(t *testing.T) {
	if os.Getenv(runnerHelperEnv) != "1" {
		t.Skip("helper process")
	}
	err := Run(os.Getenv("SDGO_RUNNER_HELPER_DIR"), Options{
		Command: os.Getenv("SDGO_RUNNER_HELPER_COMMAND"),
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
}

func writeRunnerTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func waitForServicePIDs(t *testing.T, path string, count int) []int {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil {
			lines := strings.Fields(string(data))
			if len(lines) >= count {
				pids := make([]int, 0, count)
				for _, line := range lines[:count] {
					pid, err := strconv.Atoi(line)
					if err != nil {
						t.Fatalf("parse service PID %q: %v", line, err)
					}
					pids = append(pids, pid)
				}
				return pids
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("read service PID file: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d service PIDs", count)
	return nil
}

func waitForHelperExit(t *testing.T, helper *exec.Cmd) {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- helper.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("runner helper exited with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("runner helper did not exit after interrupt")
	}
}

func waitForProcessExit(t *testing.T, pid int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if errors.Is(syscall.Kill(pid, 0), syscall.ESRCH) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("service process %d survived runner exit", pid)
}
