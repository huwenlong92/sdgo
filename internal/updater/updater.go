package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const DefaultInstallTarget = "github.com/huwenlong92/sdgo/cmd/sdgo"
const defaultModulePath = "github.com/huwenlong92/sdgo"

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

	stdout := opt.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opt.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	resolvedVersion := version
	if version == "latest" {
		latest, err := resolveLatestVersion(target)
		if err == nil && latest != "" {
			resolvedVersion = latest
			fmt.Fprintf(stdout, "latest sdgo version: %s\n", resolvedVersion)
		} else if err != nil {
			fmt.Fprintf(stdout, "latest sdgo version: unknown (%v)\n", err)
		}
	}

	path, currentVersion, currentErr := installedVersion()
	if currentErr == nil && versionsEqual(currentVersion, resolvedVersion) {
		fmt.Fprintf(stdout, "current sdgo: %s\n", path)
		fmt.Fprintf(stdout, "current sdgo version: %s\n", currentVersion)
		fmt.Fprintln(stdout, "sdgo is already up to date")
		return nil
	}

	installVersion := version
	if version == "latest" && resolvedVersion != "latest" {
		installVersion = resolvedVersion
	}
	installTarget := target + "@" + installVersion
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

func resolveLatestVersion(target string) (string, error) {
	modulePath := modulePathForTarget(target)
	out, err := exec.Command("go", "list", "-m", "-json", modulePath+"@latest").Output()
	if err != nil {
		return "", err
	}
	var info struct {
		Version string
	}
	if err := json.Unmarshal(out, &info); err != nil {
		return "", err
	}
	if info.Version == "" {
		return "", fmt.Errorf("version not found")
	}
	return info.Version, nil
}

func modulePathForTarget(target string) string {
	if target == DefaultInstallTarget {
		return defaultModulePath
	}
	if strings.HasSuffix(target, "/cmd/sdgo") {
		return strings.TrimSuffix(target, "/cmd/sdgo")
	}
	return target
}

func versionsEqual(a string, b string) bool {
	a = strings.TrimPrefix(strings.TrimSpace(a), "v")
	b = strings.TrimPrefix(strings.TrimSpace(b), "v")
	return a != "" && a == b
}

func installedVersion() (string, string, error) {
	path, err := exec.LookPath("sdgo")
	if err != nil || path == "" {
		path, err = os.Executable()
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
