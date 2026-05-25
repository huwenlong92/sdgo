package updater

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const DefaultInstallTarget = "github.com/huwenlong92/sdgo/cmd/sdgo"

type Options struct {
	Version string
	Target  string
	Stdout  io.Writer
	Stderr  io.Writer
}

func Run(opt Options) error {
	version := NormalizeVersion(opt.Version)
	target := opt.Target
	if target == "" {
		target = DefaultInstallTarget
	}
	installTarget := target + "@" + version

	stdout := opt.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opt.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	fmt.Fprintf(stdout, "updating sdgo: go install %s\n", installTarget)
	cmd := exec.Command("go", "install", installTarget)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("update sdgo: %w", err)
	}
	fmt.Fprintln(stdout, "sdgo updated")
	return nil
}

func NormalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "latest"
	}
	return version
}
