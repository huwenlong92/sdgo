package runner

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/huwenlong/sdgo/internal/project"
)

type Options struct {
	Command string
	Target  string
	Watch   string
	NoWatch bool
}

func Run(dir string, opt Options) error {
	if !project.IsGoProject(dir) {
		return fmt.Errorf("current directory is not a Go project: go.mod not found")
	}
	if opt.Command == "" {
		command, err := project.DefaultRunCommand(dir, opt.Target)
		if err != nil {
			return err
		}
		opt.Command = command
	}

	roots := watchRoots(opt.Watch)
	fmt.Fprintf(os.Stderr, "running: %s\n", opt.Command)
	if opt.NoWatch {
		fmt.Fprintln(os.Stderr, "watching: disabled")
	} else {
		fmt.Fprintf(os.Stderr, "watching: %s\n", strings.Join(roots, ", "))
	}

	watcher := NewWatcher(dir, roots)
	snapshot, err := watcher.Snapshot()
	if err != nil {
		return err
	}

	proc, err := start(dir, opt.Command)
	if err != nil {
		return err
	}

	if opt.NoWatch {
		return waitProcess(proc)
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
				fmt.Fprintf(os.Stderr, "running: %s\n", opt.Command)
				proc, err = start(dir, opt.Command)
				if err != nil {
					return err
				}
				snapshot = next
			}
		}
	}
}

func waitProcess(proc *Process) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case <-signals:
		proc.Stop()
		return nil
	case err := <-proc.done:
		return err
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

func watchRoots(raw string) []string {
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
	return []string{"."}
}
