package runner

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/huwenlong/sdgo/internal/project"
)

type Options struct {
	Command string
	Watch   string
}

func Run(dir string, opt Options) error {
	if !project.IsGoProject(dir) {
		return fmt.Errorf("current directory is not a Go project: go.mod not found")
	}
	if opt.Command == "" {
		command, err := project.DefaultRunCommand(dir)
		if err != nil {
			return err
		}
		opt.Command = command
	}

	watcher := NewWatcher(dir, watchRoots(dir, opt.Watch))
	snapshot, err := watcher.Snapshot()
	if err != nil {
		return err
	}

	proc, err := start(dir, opt.Command)
	if err != nil {
		return err
	}
	defer proc.Stop()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-signals:
			return nil
		case <-ticker.C:
			next, err := watcher.Snapshot()
			if err != nil {
				fmt.Fprintf(os.Stderr, "watch error: %v\n", err)
				continue
			}
			if !snapshot.Equal(next) {
				fmt.Fprintln(os.Stderr, "file changed, restarting...")
				proc.Stop()
				proc, err = start(dir, opt.Command)
				if err != nil {
					return err
				}
				snapshot = next
			}
		}
	}
}

type Process struct {
	cmd  *exec.Cmd
	done chan error
}

func start(dir string, command string) (*Process, error) {
	cmd := shellCommand(command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start command %q: %w", command, err)
	}

	proc := &Process{cmd: cmd, done: make(chan error, 1)}
	go func() {
		proc.done <- cmd.Wait()
	}()
	return proc, nil
}

func (p *Process) Stop() {
	if p == nil || p.cmd == nil || p.cmd.Process == nil {
		return
	}
	_ = p.cmd.Process.Signal(os.Interrupt)
	select {
	case <-p.done:
	case <-time.After(500 * time.Millisecond):
		_ = p.cmd.Process.Kill()
		<-p.done
	}
}

func shellCommand(command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/C", command)
	}
	return exec.Command("/bin/sh", "-c", command)
}

func watchRoots(dir string, raw string) []string {
	if raw != "" {
		parts := strings.Split(raw, ",")
		roots := make([]string, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				roots = append(roots, part)
			}
		}
		return roots
	}

	candidates := []string{"app", "bootstrap", "command", "configs", "core", "modules", "pkg", "cmd"}
	roots := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if _, err := os.Stat(filepath.Join(dir, candidate)); err == nil {
			roots = append(roots, candidate)
		}
	}
	if len(roots) == 0 {
		return []string{"."}
	}
	return roots
}
