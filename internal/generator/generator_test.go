package generator

import (
	"os"
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
	dir := filepath.Join(base, "template")
	files := map[string]string{
		"go.mod":              "module sdkitgo\n\ngo 1.25.0\n",
		"cmd/sdkitgo/main.go": "package main\n\nimport \"sdkitgo/command\"\n\nfunc main() { command.RegisterAll(nil) }\n",
		"bootstrap/boot.go":   "package bootstrap\n",
		"command/command.go":  "package command\n",
		"configs/config.yaml": "app:\n  name: sdkitgo\n",
		"deploy/Dockerfile":   "COPY sdkitgo/ ./\nRUN go build -o /out/sdkitgo ./cmd/sdkitgo\n",
		"tests/skip_test.go":  "package tests\n",
		".git/config":         "[core]\n",
		".air.toml":           "root = \".\"\n",
		"tmp/generated.txt":   "skip\n",
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

func contains(s string, sub string) bool {
	return strings.Contains(s, sub)
}
