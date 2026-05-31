package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateProjectCreatesProject(t *testing.T) {
	dir := t.TempDir()
	source := writeProjectSource(t, dir)

	err := GenerateProject(dir, ProjectOptions{Name: "demo", ModulePath: "github.com/acme/demo", SourceDir: source})
	if err != nil {
		t.Fatalf("GenerateProject returned error: %v", err)
	}

	for _, path := range []string{
		"go.mod",
		"cmd/demo/main.go",
		"bootstrap/boot.go",
		"command/command.go",
		"configs/config.yaml",
	} {
		if _, err := os.Stat(filepath.Join(dir, "demo", path)); err != nil {
			t.Fatalf("expected generated file %s: %v", path, err)
		}
	}

	if _, err := os.Stat(filepath.Join(dir, "demo", ".air.toml")); !os.IsNotExist(err) {
		t.Fatalf("project template should not generate .air.toml by default, got %v", err)
	}
}

func TestGenerateProjectRewritesModuleAndImports(t *testing.T) {
	dir := t.TempDir()
	source := writeProjectSource(t, dir)

	err := GenerateProject(dir, ProjectOptions{Name: "demo-app", ModulePath: "github.com/acme/demo-app", SourceDir: source})
	if err != nil {
		t.Fatalf("GenerateProject returned error: %v", err)
	}

	goMod, err := os.ReadFile(filepath.Join(dir, "demo-app", "go.mod"))
	if err != nil {
		t.Fatalf("ReadFile go.mod returned error: %v", err)
	}
	if got := string(goMod); !contains(got, "module github.com/acme/demo-app") {
		t.Fatalf("expected module path to be rewritten, got:\n%s", got)
	}

	mainGo, err := os.ReadFile(filepath.Join(dir, "demo-app", "cmd", "demoapp", "main.go"))
	if err != nil {
		t.Fatalf("ReadFile main.go returned error: %v", err)
	}
	if got := string(mainGo); !contains(got, `"github.com/acme/demo-app/command"`) {
		t.Fatalf("expected import path to be rewritten, got:\n%s", got)
	}

	dockerfile, err := os.ReadFile(filepath.Join(dir, "demo-app", "deploy", "Dockerfile"))
	if err != nil {
		t.Fatalf("ReadFile Dockerfile returned error: %v", err)
	}
	if got := string(dockerfile); !contains(got, "COPY demo-app/ ./") || !contains(got, "./cmd/demoapp") {
		t.Fatalf("expected Dockerfile paths to be rewritten, got:\n%s", got)
	}
}

func TestGenerateProjectRewritesStaleLocalImportRoot(t *testing.T) {
	dir := t.TempDir()
	source := writeProjectSource(t, dir)
	writeTemplateFile(t, source, "app/admin/provider.go", "package admin\n\nimport \"sdauth/app/infra/capability/storage\"\n\nvar _ = storage.Provider{}\n")
	writeTemplateFile(t, source, "app/infra/capability/storage/storage.go", "package storage\n\ntype Provider struct{}\n")

	err := GenerateProject(dir, ProjectOptions{Name: "demo", ModulePath: "github.com/acme/demo", SourceDir: source})
	if err != nil {
		t.Fatalf("GenerateProject returned error: %v", err)
	}

	providerGo, err := os.ReadFile(filepath.Join(dir, "demo", "app", "admin", "provider.go"))
	if err != nil {
		t.Fatalf("ReadFile provider.go returned error: %v", err)
	}
	if got := string(providerGo); !contains(got, `"github.com/acme/demo/app/infra/capability/storage"`) {
		t.Fatalf("expected stale local import root to be rewritten, got:\n%s", got)
	}
}

func TestGenerateProjectDoesNotRewriteStandardLibraryImportRoot(t *testing.T) {
	dir := t.TempDir()
	source := writeProjectSource(t, dir)
	writeTemplateFile(t, source, "app/admin/provider.go", "package admin\n\nimport \"net/http\"\n\nvar _ = http.MethodGet\n")
	writeTemplateFile(t, source, "http/local.go", "package http\n")

	err := GenerateProject(dir, ProjectOptions{Name: "demo", ModulePath: "github.com/acme/demo", SourceDir: source})
	if err != nil {
		t.Fatalf("GenerateProject returned error: %v", err)
	}

	providerGo, err := os.ReadFile(filepath.Join(dir, "demo", "app", "admin", "provider.go"))
	if err != nil {
		t.Fatalf("ReadFile provider.go returned error: %v", err)
	}
	if got := string(providerGo); !contains(got, `"net/http"`) {
		t.Fatalf("expected standard library import to be preserved, got:\n%s", got)
	}
}

func TestGenerateProjectFromAlternateTemplateIdentity(t *testing.T) {
	dir := t.TempDir()
	source := writeNamedProjectSource(t, dir, "sdadmin")

	err := GenerateProject(dir, ProjectOptions{Name: "demo", ModulePath: "mycorp/demo", SourceDir: source})
	if err != nil {
		t.Fatalf("GenerateProject returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "demo", "cmd", "demo", "main.go")); err != nil {
		t.Fatalf("expected command directory to be renamed: %v", err)
	}

	mainGo, err := os.ReadFile(filepath.Join(dir, "demo", "cmd", "demo", "main.go"))
	if err != nil {
		t.Fatalf("ReadFile main.go returned error: %v", err)
	}
	if got := string(mainGo); !contains(got, `"mycorp/demo/command"`) {
		t.Fatalf("expected alternate template import path to be rewritten, got:\n%s", got)
	}
}

func TestGenerateProjectFromMultiCommandTemplate(t *testing.T) {
	dir := t.TempDir()
	source := writeMultiCommandProjectSource(t, dir)

	err := GenerateProject(dir, ProjectOptions{Name: "demo-app", ModulePath: "github.com/acme/demo-app", SourceDir: source})
	if err != nil {
		t.Fatalf("GenerateProject returned error: %v", err)
	}

	for _, path := range []string{
		"cmd/serve/main.go",
		"cmd/api/main.go",
	} {
		if _, err := os.Stat(filepath.Join(dir, "demo-app", path)); err != nil {
			t.Fatalf("expected generated command file %s: %v", path, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "demo-app", "cmd", "demoapp", "main.go")); !os.IsNotExist(err) {
		t.Fatalf("multi-command templates should preserve command directories, got %v", err)
	}

	mainGo, err := os.ReadFile(filepath.Join(dir, "demo-app", "cmd", "serve", "main.go"))
	if err != nil {
		t.Fatalf("ReadFile main.go returned error: %v", err)
	}
	if got := string(mainGo); !contains(got, `"github.com/acme/demo-app/command"`) {
		t.Fatalf("expected import path to be rewritten, got:\n%s", got)
	}

	config, err := os.ReadFile(filepath.Join(dir, "demo-app", "configs", "config.yaml"))
	if err != nil {
		t.Fatalf("ReadFile config.yaml returned error: %v", err)
	}
	if got := string(config); !contains(got, "name: demoapp") {
		t.Fatalf("expected app name to be rewritten, got:\n%s", got)
	}

	dockerfile, err := os.ReadFile(filepath.Join(dir, "demo-app", "deploy", "Dockerfile"))
	if err != nil {
		t.Fatalf("ReadFile Dockerfile returned error: %v", err)
	}
	if got := string(dockerfile); !contains(got, "COPY demo-app/go.mod demo-app/go.sum") || !contains(got, "./cmd/serve") {
		t.Fatalf("expected Dockerfile project paths to be rewritten while preserving cmd/serve, got:\n%s", got)
	}
}

func TestGenerateProjectFromNodeTemplate(t *testing.T) {
	dir := t.TempDir()
	source := writeNodeProjectSource(t, dir)

	err := GenerateProject(dir, ProjectOptions{Name: "admin-web", SourceDir: source})
	if err != nil {
		t.Fatalf("GenerateProject returned error: %v", err)
	}

	packageJSON, err := os.ReadFile(filepath.Join(dir, "admin-web", "package.json"))
	if err != nil {
		t.Fatalf("ReadFile package.json returned error: %v", err)
	}
	if got := string(packageJSON); !contains(got, `"name": "admin-web"`) {
		t.Fatalf("expected package name to be rewritten, got:\n%s", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "admin-web", "src", "main.ts")); err != nil {
		t.Fatalf("expected frontend source file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "admin-web", "library", "build", "index.ts")); err != nil {
		t.Fatalf("expected nested frontend build source directory to be copied: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "admin-web", "src", "views", "system", "storage", "index.vue")); err != nil {
		t.Fatalf("expected nested frontend storage source directory to be copied: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "admin-web", "node_modules")); !os.IsNotExist(err) {
		t.Fatalf("node_modules should not be copied, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "admin-web", "dist")); !os.IsNotExist(err) {
		t.Fatalf("dist should not be copied, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "admin-web", "build")); !os.IsNotExist(err) {
		t.Fatalf("root build should not be copied, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "admin-web", "storage")); !os.IsNotExist(err) {
		t.Fatalf("root storage should not be copied, got %v", err)
	}
}

func TestGenerateProjectRejectsModuleForNodeTemplate(t *testing.T) {
	dir := t.TempDir()
	source := writeNodeProjectSource(t, dir)

	err := GenerateProject(dir, ProjectOptions{Name: "admin-web", ModulePath: "mycorp/admin-web", SourceDir: source})
	if err == nil {
		t.Fatalf("expected --module to be rejected for node template")
	}
	if !contains(err.Error(), "--module is only supported for Go templates") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateProjectFromGitSource(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	source := writeProjectSource(t, filepath.Join(dir, "source-base"))
	runGit(t, source, "init")
	runGit(t, source, "config", "user.email", "test@example.com")
	runGit(t, source, "config", "user.name", "Test User")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "template")

	err := GenerateProject(dir, ProjectOptions{
		Name:       "demo",
		ModulePath: "github.com/acme/demo",
		SourceDir:  "file://" + filepath.ToSlash(source),
	})
	if err != nil {
		t.Fatalf("GenerateProject returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "demo", "cmd", "demo", "main.go")); err != nil {
		t.Fatalf("expected generated file from git source: %v", err)
	}
}

func TestGenerateProjectFromGitSourceBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	source := writeProjectSource(t, filepath.Join(dir, "source-base"))
	runGit(t, source, "init")
	runGit(t, source, "config", "user.email", "test@example.com")
	runGit(t, source, "config", "user.name", "Test User")
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "template")
	runGit(t, source, "checkout", "-b", "feature-template")
	if err := os.WriteFile(filepath.Join(source, "configs", "branch.yaml"), []byte("from: branch\n"), 0o644); err != nil {
		t.Fatalf("WriteFile branch marker returned error: %v", err)
	}
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "branch template")

	err := GenerateProject(dir, ProjectOptions{
		Name:       "demo",
		ModulePath: "github.com/acme/demo",
		SourceDir:  "file://" + filepath.ToSlash(source),
		Branch:     "feature-template",
	})
	if err != nil {
		t.Fatalf("GenerateProject returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "demo", "configs", "branch.yaml")); err != nil {
		t.Fatalf("expected generated file from git branch: %v", err)
	}
}

func TestGenerateModuleCreatesModule(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir)

	err := GenerateModule(dir, ModuleOptions{Name: "user"})
	if err != nil {
		t.Fatalf("GenerateModule returned error: %v", err)
	}

	for _, path := range []string{
		"modules/user/module.go",
		"modules/user/route.go",
		"modules/user/handler.go",
		"modules/user/service.go",
		"modules/user/repository.go",
		"modules/user/model.go",
	} {
		if _, err := os.Stat(filepath.Join(dir, path)); err != nil {
			t.Fatalf("expected generated file %s: %v", path, err)
		}
	}
}

func TestGenerateModuleCreatesDocs(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir)

	err := GenerateModule(dir, ModuleOptions{Name: "user-profile"})
	if err != nil {
		t.Fatalf("GenerateModule returned error: %v", err)
	}

	for _, path := range []string{
		"docs/modules/user_profile.md",
		"docs/usage/user_profile.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, path)); err != nil {
			t.Fatalf("expected generated doc %s: %v", path, err)
		}
	}
}

func writeGoMod(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testapp\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("WriteFile go.mod returned error: %v", err)
	}
}

func writeProjectSource(t *testing.T, base string) string {
	t.Helper()
	return writeNamedProjectSource(t, base, "sdkitgo")
}

func writeNamedProjectSource(t *testing.T, base string, sourceName string) string {
	t.Helper()
	dir := filepath.Join(base, "template")
	files := map[string]string{
		"go.mod":                         "module " + sourceName + "\n\ngo 1.25.0\n",
		"cmd/" + sourceName + "/main.go": "package main\n\nimport \"" + sourceName + "/command\"\n\nfunc main() { command.RegisterAll(nil) }\n",
		"bootstrap/boot.go":              "package bootstrap\n",
		"command/command.go":             "package command\n",
		"configs/config.yaml":            "app:\n  name: " + sourceName + "\n",
		"deploy/Dockerfile":              "COPY " + sourceName + "/ ./\nRUN go build -o /out/" + sourceName + " ./cmd/" + sourceName + "\n",
		"tests/skip_test.go":             "package tests\n",
		".git/config":                    "[core]\n",
		".air.toml":                      "root = \".\"\n",
		"tmp/generated.txt":              "skip\n",
	}
	for path, content := range files {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("MkdirAll %s returned error: %v", path, err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile %s returned error: %v", path, err)
		}
	}
	if err := os.Symlink(filepath.Join(base, "external"), filepath.Join(dir, "_reference")); err != nil {
		t.Fatalf("Symlink returned error: %v", err)
	}
	return dir
}

func writeTemplateFile(t *testing.T, dir string, path string, content string) {
	t.Helper()
	full := filepath.Join(dir, path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("MkdirAll %s returned error: %v", path, err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile %s returned error: %v", path, err)
	}
}

func writeMultiCommandProjectSource(t *testing.T, base string) string {
	t.Helper()
	dir := filepath.Join(base, "template")
	files := map[string]string{
		"go.mod":               "module sdkitgo\n\ngo 1.25.0\n",
		"cmd/serve/main.go":    "package main\n\nimport \"sdkitgo/command\"\n\nfunc main() { command.RegisterAll(nil) }\n",
		"cmd/api/main.go":      "package main\n\nimport \"sdkitgo/command\"\n\nfunc main() { command.RegisterAll(nil) }\n",
		"bootstrap/boot.go":    "package bootstrap\n",
		"command/command.go":   "package command\n",
		"configs/config.yaml":  "app:\n  name: sdkitgo\n",
		"deploy/Dockerfile":    "COPY sdkitgo/go.mod sdkitgo/go.sum ./\nCOPY sdkitgo/ ./\nRUN go build -o /out/serve ./cmd/serve\n",
		"deploy/ignore.txt":    "sdkitgo\n",
		"tests/skip_test.go":   "package tests\n",
		"tmp/generated.txt":    "skip\n",
		"vendor/skip/skip.txt": "skip\n",
	}
	for path, content := range files {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("MkdirAll %s returned error: %v", path, err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile %s returned error: %v", path, err)
		}
	}
	return dir
}

func writeNodeProjectSource(t *testing.T, base string) string {
	t.Helper()
	dir := filepath.Join(base, "template")
	files := map[string]string{
		"package.json":                       "{\n  \"name\": \"admin-template\",\n  \"version\": \"0.1.0\",\n  \"scripts\": {\"dev\": \"vite\"}\n}\n",
		"package-lock.json":                  "{\n  \"name\": \"admin-template\",\n  \"packages\": {\"\": {\"name\": \"admin-template\"}}\n}\n",
		"index.html":                         "<title>admin-template</title>\n",
		"src/main.ts":                        "console.log('hello')\n",
		"library/build/index.ts":             "export const createVitePlugin = () => []\n",
		"src/views/system/storage/index.vue": "<template><div /></template>\n",
		"node_modules/vite/index":            "skip\n",
		"dist/assets/index.js":               "skip\n",
		"build/cache.txt":                    "skip\n",
		"storage/runtime.txt":                "skip\n",
		".vite/deps/package.json":            "skip\n",
		"coverage/lcov.info":                 "skip\n",
		".git/config":                        "[core]\n",
		"tmp/generated.txt":                  "skip\n",
	}
	for path, content := range files {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("MkdirAll %s returned error: %v", path, err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile %s returned error: %v", path, err)
		}
	}
	return dir
}

func contains(s string, sub string) bool {
	return strings.Contains(s, sub)
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v returned error: %v\n%s", args, err, out)
	}
}
