package project

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func IsGoProject(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "go.mod"))
	return err == nil
}

func DefaultRunCommand(dir string, target string) (string, error) {
	cmdDir, err := defaultCommandDir(dir, target)
	if err != nil {
		return "", err
	}
	return "go run ./" + filepath.ToSlash(cmdDir), nil
}

func defaultCommandDir(dir string, target string) (string, error) {
	target = strings.TrimSpace(target)
	if target != "" {
		cmdDir := filepath.Join("cmd", target)
		if hasMainGo(filepath.Join(dir, cmdDir)) {
			return cmdDir, nil
		}
		return "", fmt.Errorf("cmd/%s/main.go not found, pass --cmd", target)
	}

	candidates, err := commandDirs(dir)
	if err != nil {
		return "", err
	}
	for _, candidate := range candidates {
		if filepath.Base(candidate) == "serve" {
			return candidate, nil
		}
	}
	switch len(candidates) {
	case 0:
		return "", fmt.Errorf("cmd/*/main.go not found, pass --cmd")
	case 1:
		return candidates[0], nil
	default:
		return "", fmt.Errorf("multiple cmd entries found, pass target or --cmd: %v", candidates)
	}
}

func commandDirs(dir string) ([]string, error) {
	root := filepath.Join(dir, "cmd")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("cmd directory not found, pass --cmd")
		}
		return nil, fmt.Errorf("read cmd directory: %w", err)
	}
	var candidates []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if hasMainGo(filepath.Join(root, entry.Name())) {
			candidates = append(candidates, filepath.Join("cmd", entry.Name()))
		}
	}
	sort.Strings(candidates)
	return candidates, nil
}

func hasMainGo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "main.go"))
	return err == nil
}
