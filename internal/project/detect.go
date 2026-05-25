package project

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func IsGoProject(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "go.mod"))
	return err == nil
}

func DefaultRunCommand(dir string) (string, error) {
	cmdDir, err := defaultCommandDir(dir)
	if err != nil {
		return "", err
	}
	return "go run ./" + filepath.ToSlash(cmdDir) + " serve", nil
}

func defaultCommandDir(dir string) (string, error) {
	root := filepath.Join(dir, "cmd")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("cmd directory not found, pass --cmd")
		}
		return "", fmt.Errorf("read cmd directory: %w", err)
	}
	var candidates []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		mainPath := filepath.Join(root, entry.Name(), "main.go")
		if _, err := os.Stat(mainPath); err == nil {
			candidates = append(candidates, filepath.Join("cmd", entry.Name()))
		}
	}
	sort.Strings(candidates)
	switch len(candidates) {
	case 0:
		return "", fmt.Errorf("cmd/*/main.go not found, pass --cmd")
	case 1:
		return candidates[0], nil
	default:
		return "", fmt.Errorf("multiple cmd entries found, pass --cmd: %v", candidates)
	}
}
