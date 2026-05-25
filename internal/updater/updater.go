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

	fmt.Fprintf(stdout, "upgrading sdgo: go install %s\n", installTarget)
	cmd := exec.Command("go", "install", installTarget)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("upgrade sdgo: %w", err)
	}
	fmt.Fprintln(stdout, "sdgo upgraded")
	path, version, err := installedVersion()
	if err != nil {
		fmt.Fprintf(stdout, "current sdgo version: unknown (%v)\n", err)
		return nil
	}
	fmt.Fprintf(stdout, "current sdgo: %s\n", path)
	fmt.Fprintf(stdout, "current sdgo version: %s\n", version)
	return nil
}

func NormalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "latest"
	}
	return version
}

func installedVersion() (string, string, error) {
	path, err := os.Executable()
	if err != nil || path == "" {
		path, err = exec.LookPath("sdgo")
		if err != nil {
			return "", "", err
		}
	}

	out, err := exec.Command(path, "version").Output()
	if err != nil {
		return path, "", err
	}
	return path, strings.TrimSpace(string(out)), nil
}
