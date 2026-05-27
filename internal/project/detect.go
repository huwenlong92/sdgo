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
	tags, err := BuildTags(dir)
	if err != nil {
		return "", err
	}
	args := []string{"go", "run"}
	if len(tags) > 0 {
		args = append(args, "-tags", strings.Join(tags, ","))
	}
	args = append(args, "./"+filepath.ToSlash(cmdDir))
	return strings.Join(args, " "), nil
}

func BuildTags(dir string) ([]string, error) {
	path := filepath.Join(dir, "build.yaml")
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read build.yaml: %w", err)
	}
	tags, err := parseBuildTags(string(content))
	if err != nil {
		return nil, fmt.Errorf("parse build.yaml: %w", err)
	}
	return tags, nil
}

func parseBuildTags(content string) ([]string, error) {
	seen := make(map[string]struct{})
	inTags := false
	tagsIndent := -1
	for _, line := range strings.Split(content, "\n") {
		raw := stripComment(line)
		if strings.TrimSpace(raw) == "" {
			continue
		}
		trimmed := strings.TrimSpace(raw)
		indent := leadingSpaces(raw)
		if !inTags {
			if strings.HasPrefix(trimmed, "tags:") {
				inTags = true
				tagsIndent = indent
				addInlineTags(seen, strings.TrimSpace(strings.TrimPrefix(trimmed, "tags:")))
			}
			continue
		}
		if indent <= tagsIndent && !strings.HasPrefix(trimmed, "-") {
			break
		}
		switch {
		case strings.HasPrefix(trimmed, "-"):
			addTag(seen, strings.TrimSpace(strings.TrimPrefix(trimmed, "-")))
		case strings.Contains(trimmed, ":"):
			parts := strings.SplitN(trimmed, ":", 2)
			if tagEnabled(strings.TrimSpace(parts[1])) {
				addTag(seen, strings.TrimSpace(parts[0]))
			}
		default:
			addInlineTags(seen, trimmed)
		}
	}
	tags := make([]string, 0, len(seen))
	for tag := range seen {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags, nil
}

func stripComment(line string) string {
	if idx := strings.Index(line, "#"); idx >= 0 {
		return line[:idx]
	}
	return line
}

func leadingSpaces(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

func addInlineTags(seen map[string]struct{}, value string) {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	for _, part := range strings.Split(value, ",") {
		addTag(seen, part)
	}
}

func addTag(seen map[string]struct{}, tag string) {
	tag = strings.Trim(strings.TrimSpace(tag), `"'`)
	if tag != "" {
		seen[tag] = struct{}{}
	}
}

func tagEnabled(value string) bool {
	value = strings.ToLower(strings.Trim(strings.TrimSpace(value), `"'`))
	switch value {
	case "true", "yes", "on", "1":
		return true
	default:
		return false
	}
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
