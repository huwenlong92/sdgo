package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ProjectOptions struct {
	Name       string
	ModulePath string
	SourceDir  string
	Force      bool
}

func GenerateProject(baseDir string, opt ProjectOptions) error {
	data, err := normalizeName(opt.Name)
	if err != nil {
		return err
	}
	data.ProjectName = opt.Name
	if opt.ModulePath == "" {
		data.ModulePath = opt.Name
	} else {
		data.ModulePath = opt.ModulePath
	}
	data.Year = time.Now().Year()

	sourceDir, err := resolveSourceDir(baseDir, opt.SourceDir)
	if err != nil {
		return err
	}

	dst := filepath.Join(baseDir, opt.Name)
	if info, err := os.Stat(dst); err == nil && !info.IsDir() {
		return fmt.Errorf("target exists and is not a directory: %s", dst)
	} else if err == nil && !opt.Force {
		return fmt.Errorf("target directory already exists: %s, use --force to overwrite", dst)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat target directory: %w", err)
	}

	return copyProject(sourceDir, dst, data, opt.Force)
}

func resolveSourceDir(baseDir string, explicit string) (string, error) {
	candidates := []string{}
	if explicit != "" {
		candidates = append(candidates, explicit)
	}
	if env := os.Getenv("SDKITGO_SOURCE"); env != "" {
		candidates = append(candidates, env)
	}
	candidates = append(candidates,
		baseDir,
		filepath.Join(baseDir, "sdkitgo"),
		filepath.Join(baseDir, "..", "sdkitgo"),
		"/Users/huwenlong/data/lab/sdkitgo",
	)

	for _, candidate := range candidates {
		path, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if isSdkitgoSource(path) {
			return path, nil
		}
	}
	return "", fmt.Errorf("sdkitgo source project not found, pass --source or set SDKITGO_SOURCE")
}

func isSdkitgoSource(dir string) bool {
	for _, path := range []string{
		filepath.Join(dir, "go.mod"),
		filepath.Join(dir, "cmd", "sdkitgo", "main.go"),
		filepath.Join(dir, "bootstrap"),
		filepath.Join(dir, "command"),
	} {
		if _, err := os.Stat(path); err != nil {
			return false
		}
	}
	return true
}

func copyProject(src string, dst string, data TemplateData, force bool) error {
	return filepath.WalkDir(src, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if entry.IsDir() {
			if shouldSkipProjectPath(rel, true) {
				return filepath.SkipDir
			}
			return os.MkdirAll(mappedProjectPath(dst, rel, data), 0o755)
		}
		if shouldSkipProjectPath(rel, false) {
			return nil
		}
		return copyProjectFile(path, mappedProjectPath(dst, rel, data), rel, data, force)
	})
}

func mappedProjectPath(dst string, rel string, data TemplateData) string {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) >= 2 && parts[0] == "cmd" && parts[1] == "sdkitgo" {
		parts[1] = data.LowerName
	}
	return filepath.Join(append([]string{dst}, parts...)...)
}

func copyProjectFile(src string, dst string, rel string, data TemplateData, force bool) error {
	if !force {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("file already exists: %s, use --force to overwrite", dst)
		} else if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("stat file: %w", err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat source file: %w", err)
	}
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read source file %s: %w", src, err)
	}
	content = rewriteProjectFile(rel, content, data)
	if err := os.WriteFile(dst, content, info.Mode().Perm()); err != nil {
		return fmt.Errorf("write file %s: %w", dst, err)
	}
	return nil
}

func rewriteProjectFile(rel string, content []byte, data TemplateData) []byte {
	name := filepath.Base(rel)
	if filepath.ToSlash(rel) == "go.mod" {
		lines := bytes.Split(content, []byte("\n"))
		if len(lines) > 0 && bytes.HasPrefix(lines[0], []byte("module ")) {
			lines[0] = []byte("module " + data.ModulePath)
			content = bytes.Join(lines, []byte("\n"))
		}
	}
	out := string(content)
	if filepath.Ext(name) == ".go" {
		replacer := strings.NewReplacer(
			`"sdkitgo/`, `"`+data.ModulePath+`/`,
			"`sdkitgo/", "`"+data.ModulePath+"/",
			`"sdkitgo"`, `"`+data.LowerName+`"`,
			"sdkitgo ", data.LowerName+" ",
			`"sdkitgo:`, `"`+data.LowerName+`:`,
			"@sdkitgo.com", "@"+data.LowerName+".com",
		)
		out = replacer.Replace(out)
	}
	if filepath.Ext(name) == ".md" {
		replacer := strings.NewReplacer(
			"sdkitgo/", data.ModulePath+"/",
			"/sdkitgo/", "/"+data.LowerName+"/",
			"`sdkitgo`", "`"+data.LowerName+"`",
			"`sdkitgo ", "`"+data.LowerName+" ",
			" sdkitgo ", " "+data.LowerName+" ",
		)
		out = replacer.Replace(out)
	}
	out = strings.ReplaceAll(out, "cmd/sdkitgo", "cmd/"+data.LowerName)
	if name == "Dockerfile" {
		out = strings.ReplaceAll(out, "sdkitgo", data.LowerName)
		out = strings.ReplaceAll(out, "COPY "+data.LowerName+"/go.mod "+data.LowerName+"/go.sum", "COPY "+data.ProjectName+"/go.mod "+data.ProjectName+"/go.sum")
		out = strings.ReplaceAll(out, "COPY "+data.LowerName+"/ ./", "COPY "+data.ProjectName+"/ ./")
	}
	if shouldRewriteProjectNameInText(rel) {
		out = strings.ReplaceAll(out, "sdkitgo", data.LowerName)
	}
	if name == "Dockerfile.dockerignore" {
		out = strings.ReplaceAll(out, data.LowerName+"/", data.ProjectName+"/")
	}
	if filepath.ToSlash(rel) == "configs/config.yaml" {
		out = strings.ReplaceAll(out, "name: sdkitgo", "name: "+data.LowerName)
	}
	return []byte(out)
}

func shouldRewriteProjectNameInText(rel string) bool {
	name := filepath.Base(rel)
	ext := filepath.Ext(name)
	switch ext {
	case ".yaml", ".yml", ".toml", ".json", ".sh":
		return true
	}
	switch name {
	case "Makefile", "Dockerfile.dockerignore":
		return true
	default:
		return false
	}
}

func shouldSkipProjectPath(rel string, isDir bool) bool {
	slash := filepath.ToSlash(rel)
	name := filepath.Base(slash)
	if isDir {
		switch name {
		case ".git", ".gitnexus", ".cache", ".claude", ".codex", ".idea", ".vscode", "bin", "dist", "logs", "node_modules", "storage", "tests", "tmp", ".tmp", "vendor":
			return true
		}
	}
	switch name {
	case ".DS_Store", ".air.toml", "AGENTS.md", "CLAUDE.md", "tips.md":
		return true
	default:
		return false
	}
}
