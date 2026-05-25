package generator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ProjectOptions struct {
	Name       string
	ModulePath string
	SourceDir  string
	Template   string
	Branch     string
	Force      bool
}

type sourceIdentity struct {
	Kind        templateKind
	ModulePath  string
	CommandName string
	PackageName string
}

type templateKind string

const (
	templateKindGo   templateKind = "go"
	templateKindNode templateKind = "node"
)

type TemplateInfo struct {
	Name        string
	Kind        string
	Description string
	Sources     []string
	Default     bool
}

func BuiltinTemplates() []TemplateInfo {
	return []TemplateInfo{
		{
			Name:        "sdkitgo",
			Kind:        string(templateKindGo),
			Description: "Default sdkitgo Go backend template.",
			Sources:     []string{"git@gitee.com:sd0/sdkitgo.git"},
			Default:     true,
		},
		{
			Name:        "sdkitgo-admin-vue",
			Kind:        string(templateKindNode),
			Description: "sdkitgo admin Vue frontend template.",
			Sources:     []string{"git@gitee.com:sd0/admin.sdkitgo.cn.git"},
		},
		{
			Name:        "sdkitgo-portal-vue",
			Kind:        string(templateKindNode),
			Description: "sdkitgo portal Vue frontend template.",
			Sources:     []string{"git@gitee.com:sd0/portal.sdkit.cn.git"},
		},
	}
}

func BuiltinTemplate(name string) (TemplateInfo, bool) {
	name = normalizeTemplateName(name)
	for _, template := range BuiltinTemplates() {
		if template.Name == name {
			return template, true
		}
	}
	return TemplateInfo{}, false
}

func TemplateSourceEnv(name string) string {
	return templateSourceEnv(normalizeTemplateName(name))
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

	sourceDir, source, cleanup, err := resolveSourceDir(baseDir, opt)
	if err != nil {
		return err
	}
	defer cleanup()
	if source.Kind != templateKindGo && opt.ModulePath != "" {
		return fmt.Errorf("--module is only supported for Go templates")
	}

	dst := filepath.Join(baseDir, opt.Name)
	if info, err := os.Stat(dst); err == nil && !info.IsDir() {
		return fmt.Errorf("target exists and is not a directory: %s", dst)
	} else if err == nil && !opt.Force {
		return fmt.Errorf("target directory already exists: %s, use --force to overwrite", dst)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat target directory: %w", err)
	}

	return copyProject(sourceDir, dst, data, source, opt.Force)
}

func resolveSourceDir(baseDir string, opt ProjectOptions) (string, sourceIdentity, func(), error) {
	template := normalizeTemplateName(opt.Template)
	candidates := []string{}
	if opt.SourceDir != "" {
		candidates = append(candidates, opt.SourceDir)
	}
	if env := os.Getenv(templateSourceEnv(template)); env != "" {
		candidates = append(candidates, env)
	}
	if template == "sdkitgo" {
		if env := os.Getenv("SDKITGO_SOURCE"); env != "" {
			candidates = append(candidates, env)
		}
	}
	if env := os.Getenv("SDGO_TEMPLATE_SOURCE"); env != "" {
		candidates = append(candidates, env)
	}
	candidates = append(candidates,
		baseDir,
		filepath.Join(baseDir, template),
		filepath.Join(baseDir, "..", template),
	)
	candidates = append(candidates, builtinTemplateSources(template)...)

	for _, candidate := range candidates {
		if isGitSource(candidate) {
			path, cleanup, err := cloneGitSource(candidate, opt.Branch)
			if err != nil {
				return "", sourceIdentity{}, nil, err
			}
			source, ok, err := inspectTemplateSource(path)
			if err != nil {
				cleanup()
				return "", sourceIdentity{}, nil, err
			}
			if ok {
				return path, source, cleanup, nil
			}
			cleanup()
			return "", sourceIdentity{}, nil, fmt.Errorf("git source is not a valid sdgo template: %s", candidate)
		}
		path, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		source, ok, err := inspectTemplateSource(path)
		if err != nil {
			return "", sourceIdentity{}, nil, err
		}
		if ok {
			return path, source, func() {}, nil
		}
	}
	return "", sourceIdentity{}, nil, fmt.Errorf("template source %q not found, pass --source or set %s", template, templateSourceEnv(template))
}

func normalizeTemplateName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "sdkitgo"
	}
	return name
}

func builtinTemplateSources(template string) []string {
	info, ok := BuiltinTemplate(template)
	if !ok {
		return nil
	}
	return append([]string(nil), info.Sources...)
}

func templateSourceEnv(template string) string {
	var b strings.Builder
	b.WriteString("SDGO_TEMPLATE_")
	for _, r := range strings.ToUpper(template) {
		if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('_')
	}
	b.WriteString("_SOURCE")
	return b.String()
}

func isGitSource(source string) bool {
	source = strings.TrimSpace(strings.ToLower(source))
	if strings.HasPrefix(source, "git@") {
		return true
	}
	for _, prefix := range []string{"https://", "http://", "ssh://", "git://", "file://"} {
		if strings.HasPrefix(source, prefix) {
			return true
		}
	}
	return false
}

func cloneGitSource(source string, branch string) (string, func(), error) {
	dir, err := os.MkdirTemp("", "sdgo-template-*")
	if err != nil {
		return "", nil, fmt.Errorf("create temp template directory: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(dir) }

	args := []string{"clone", "--depth=1"}
	if branch = strings.TrimSpace(branch); branch != "" {
		args = append(args, "--branch", branch, "--single-branch")
	}
	args = append(args, source, dir)

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		cleanup()
		return "", nil, newCloneError(source, err)
	}
	return dir, cleanup, nil
}

func newCloneError(source string, err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return fmt.Errorf("failed to clone template source: %s\n\nGit could not access the repository or ref.\nCheck that:\n- your SSH key or credential helper can access the repository\n- you have permission to the template repository\n- the repository URL is correct\n- the requested branch or tag exists\n\nYou can also pass a local template source:\n  sdgo new <project> --source ../sdkitgo\n\nOriginal error: %w", source, err)
	}
	return fmt.Errorf("failed to clone template source %q: %w", source, err)
}

func inspectTemplateSource(dir string) (sourceIdentity, bool, error) {
	if source, ok, err := inspectGoTemplateSource(dir); ok || err != nil {
		return source, ok, err
	}
	if source, ok, err := inspectNodeTemplateSource(dir); ok || err != nil {
		return source, ok, err
	}
	return sourceIdentity{}, false, nil
}

func inspectGoTemplateSource(dir string) (sourceIdentity, bool, error) {
	for _, path := range []string{
		filepath.Join(dir, "go.mod"),
		filepath.Join(dir, "bootstrap"),
		filepath.Join(dir, "command"),
	} {
		if _, err := os.Stat(path); err != nil {
			return sourceIdentity{}, false, nil
		}
	}
	modulePath, err := readGoModulePath(filepath.Join(dir, "go.mod"))
	if err != nil {
		return sourceIdentity{}, false, err
	}
	commandName, err := singleCommandName(dir)
	if err != nil {
		return sourceIdentity{}, false, err
	}
	if commandName == "" {
		return sourceIdentity{}, false, nil
	}
	return sourceIdentity{Kind: templateKindGo, ModulePath: modulePath, CommandName: commandName}, true, nil
}

func inspectNodeTemplateSource(dir string) (sourceIdentity, bool, error) {
	path := filepath.Join(dir, "package.json")
	if _, err := os.Stat(path); err != nil {
		return sourceIdentity{}, false, nil
	}
	packageName, err := readPackageName(path)
	if err != nil {
		return sourceIdentity{}, false, err
	}
	return sourceIdentity{Kind: templateKindNode, PackageName: packageName}, true, nil
}

func readGoModulePath(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			modulePath := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			if modulePath != "" {
				return modulePath, nil
			}
		}
	}
	return "", fmt.Errorf("module path not found in go.mod")
}

func readPackageName(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read package.json: %w", err)
	}
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(content, &data); err != nil {
		return "", fmt.Errorf("parse package.json: %w", err)
	}
	if strings.TrimSpace(data.Name) == "" {
		return "", fmt.Errorf("package name not found in package.json")
	}
	return data.Name, nil
}

func singleCommandName(dir string) (string, error) {
	root := filepath.Join(dir, "cmd")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read cmd directory: %w", err)
	}

	var candidates []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, entry.Name(), "main.go")); err == nil {
			candidates = append(candidates, entry.Name())
		}
	}
	switch len(candidates) {
	case 0:
		return "", nil
	case 1:
		return candidates[0], nil
	default:
		return "", fmt.Errorf("multiple template command entries found: %v", candidates)
	}
}

func copyProject(src string, dst string, data TemplateData, source sourceIdentity, force bool) error {
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
			return os.MkdirAll(mappedProjectPath(dst, rel, data, source), 0o755)
		}
		if shouldSkipProjectPath(rel, false) {
			return nil
		}
		return copyProjectFile(path, mappedProjectPath(dst, rel, data, source), rel, data, source, force)
	})
}

func mappedProjectPath(dst string, rel string, data TemplateData, source sourceIdentity) string {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) >= 2 && parts[0] == "cmd" && parts[1] == source.CommandName {
		parts[1] = data.LowerName
	}
	return filepath.Join(append([]string{dst}, parts...)...)
}

func copyProjectFile(src string, dst string, rel string, data TemplateData, source sourceIdentity, force bool) error {
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
	content = rewriteProjectFile(rel, content, data, source)
	if err := os.WriteFile(dst, content, info.Mode().Perm()); err != nil {
		return fmt.Errorf("write file %s: %w", dst, err)
	}
	return nil
}

func rewriteProjectFile(rel string, content []byte, data TemplateData, source sourceIdentity) []byte {
	switch source.Kind {
	case templateKindGo:
		return rewriteGoProjectFile(rel, content, data, source)
	case templateKindNode:
		return rewriteNodeProjectFile(rel, content, data, source)
	default:
		return content
	}
}

func rewriteGoProjectFile(rel string, content []byte, data TemplateData, source sourceIdentity) []byte {
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
			`"`+source.ModulePath+`/`, `"`+data.ModulePath+`/`,
			"`"+source.ModulePath+"/", "`"+data.ModulePath+"/",
			`"`+source.CommandName+`"`, `"`+data.LowerName+`"`,
			source.CommandName+" ", data.LowerName+" ",
			`"`+source.CommandName+`:`, `"`+data.LowerName+`:`,
			"@"+source.CommandName+".com", "@"+data.LowerName+".com",
		)
		out = replacer.Replace(out)
	}
	if filepath.Ext(name) == ".md" {
		replacer := strings.NewReplacer(
			source.ModulePath+"/", data.ModulePath+"/",
			"/"+source.CommandName+"/", "/"+data.LowerName+"/",
			"`"+source.CommandName+"`", "`"+data.LowerName+"`",
			"`"+source.CommandName+" ", "`"+data.LowerName+" ",
			" "+source.CommandName+" ", " "+data.LowerName+" ",
		)
		out = replacer.Replace(out)
	}
	out = strings.ReplaceAll(out, "cmd/"+source.CommandName, "cmd/"+data.LowerName)
	if name == "Dockerfile" {
		out = strings.ReplaceAll(out, source.CommandName, data.LowerName)
		out = strings.ReplaceAll(out, "COPY "+data.LowerName+"/go.mod "+data.LowerName+"/go.sum", "COPY "+data.ProjectName+"/go.mod "+data.ProjectName+"/go.sum")
		out = strings.ReplaceAll(out, "COPY "+data.LowerName+"/ ./", "COPY "+data.ProjectName+"/ ./")
	}
	if shouldRewriteProjectNameInText(rel) {
		out = strings.ReplaceAll(out, source.CommandName, data.LowerName)
	}
	if name == "Dockerfile.dockerignore" {
		out = strings.ReplaceAll(out, data.LowerName+"/", data.ProjectName+"/")
	}
	if filepath.ToSlash(rel) == "configs/config.yaml" {
		out = strings.ReplaceAll(out, "name: "+source.CommandName, "name: "+data.LowerName)
	}
	return []byte(out)
}

func rewriteNodeProjectFile(rel string, content []byte, data TemplateData, source sourceIdentity) []byte {
	name := filepath.Base(rel)
	slash := filepath.ToSlash(rel)
	if name == "package.json" || name == "package-lock.json" {
		var value any
		if err := json.Unmarshal(content, &value); err == nil {
			if rewritePackageJSONName(value, source.PackageName, data.ProjectName) {
				if out, err := json.MarshalIndent(value, "", "  "); err == nil {
					return append(out, '\n')
				}
			}
		}
	}
	if slash == "index.html" || filepath.Ext(name) == ".md" {
		return []byte(strings.ReplaceAll(string(content), source.PackageName, data.ProjectName))
	}
	return content
}

func rewritePackageJSONName(value any, oldName string, newName string) bool {
	obj, ok := value.(map[string]any)
	if !ok {
		return false
	}
	changed := false
	if name, ok := obj["name"].(string); ok && name == oldName {
		obj["name"] = newName
		changed = true
	}
	if packages, ok := obj["packages"].(map[string]any); ok {
		if root, ok := packages[""].(map[string]any); ok {
			if name, ok := root["name"].(string); ok && name == oldName {
				root["name"] = newName
				changed = true
			}
		}
	}
	return changed
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
		case ".git", ".gitnexus", ".cache", ".claude", ".codex", ".idea", ".next", ".nuxt", ".svelte-kit", ".vite", ".vscode", "bin", "build", "coverage", "dist", "logs", "node_modules", "storage", "tests", "tmp", ".tmp", "vendor":
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
