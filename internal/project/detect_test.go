package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsGoProject(t *testing.T) {
	dir := t.TempDir()
	if IsGoProject(dir) {
		t.Fatalf("expected directory without go.mod to be false")
	}

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module demo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	if !IsGoProject(dir) {
		t.Fatalf("expected directory with go.mod to be true")
	}
}

func TestDefaultRunCommand(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "cmd", "demo", "main.go")
	if err := os.MkdirAll(filepath.Dir(mainPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(mainPath, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	got, err := DefaultRunCommand(dir)
	if err != nil {
		t.Fatalf("DefaultRunCommand returned error: %v", err)
	}
	if got != "go run ./cmd/demo serve" {
		t.Fatalf("unexpected command: %s", got)
	}
}

func TestDefaultRunCommandRequiresExplicitCommandWhenAmbiguous(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"api", "worker"} {
		mainPath := filepath.Join(dir, "cmd", name, "main.go")
		if err := os.MkdirAll(filepath.Dir(mainPath), 0o755); err != nil {
			t.Fatalf("MkdirAll returned error: %v", err)
		}
		if err := os.WriteFile(mainPath, []byte("package main\n"), 0o644); err != nil {
			t.Fatalf("WriteFile returned error: %v", err)
		}
	}

	if _, err := DefaultRunCommand(dir); err == nil {
		t.Fatalf("expected ambiguous cmd entries to return an error")
	}
}
