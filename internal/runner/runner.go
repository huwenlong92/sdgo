package runner

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/huwenlong92/sdgo/internal/project"
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
	printStartup(opt, roots)

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
	signal.Notify(signals, shutdownSignals()...)
	defer signal.Stop(signals)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-signals:
			return nil
		case err := <-proc.done:
			return err
		case <-ticker.C:
			next, err := watcher.Snapshot()
			if err != nil {
				fmt.Fprintf(os.Stderr, "watch error: %v\n", err)
				continue
			}
			if !snapshot.Equal(next) {
				printRestart(opt.Command)
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

func printStartup(opt Options, roots []string) {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "+------------------------------------------------------------+")
	fmt.Fprintln(os.Stderr, "| sdgo                                                       |")
	fmt.Fprintln(os.Stderr, "+------------------------------------------------------------+")
	fmt.Fprintln(os.Stderr, "  status    starting")
	fmt.Fprintf(os.Stderr, "  command   %s\n", opt.Command)
	if opt.NoWatch {
		fmt.Fprintln(os.Stderr, "  watch     disabled")
	} else {
		fmt.Fprintln(os.Stderr, "  watch     enabled")
		fmt.Fprintf(os.Stderr, "  paths     %s\n", strings.Join(roots, ", "))
		fmt.Fprintln(os.Stderr, "  ignores   .git, node_modules, vendor, cache/temp/runtime/output dirs, *_test.go")
	}
	fmt.Fprintln(os.Stderr, "+------------------------------------------------------------+")
	fmt.Fprintln(os.Stderr, "")
}

func printRestart(command string) {
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "+------------------------------------------------------------+")
	fmt.Fprintln(os.Stderr, "| sdgo                                                       |")
	fmt.Fprintln(os.Stderr, "+------------------------------------------------------------+")
	fmt.Fprintln(os.Stderr, "  status    file changed, restarting")
	fmt.Fprintf(os.Stderr, "  command   %s\n", command)
	fmt.Fprintln(os.Stderr, "+------------------------------------------------------------+")
	fmt.Fprintln(os.Stderr)
}

func waitProcess(proc *Process) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, shutdownSignals()...)
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
	configureCommand(cmd)

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
	select {
	case <-p.done:
		return
	default:
	}
	_ = interruptCommand(p.cmd)
	select {
	case <-p.done:
	case <-time.After(2 * time.Second):
		_ = killCommand(p.cmd)
		select {
		case <-p.done:
		case <-time.After(time.Second):
		}
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
