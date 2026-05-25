package renderer

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRenderTemplateSuccess(t *testing.T) {
	out, err := Render("hello {{ .Name }}", map[string]string{"Name": "sdkitgo"})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	if string(out) != "hello sdkitgo" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRenderToFileReplacesVariables(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "out.txt")

	err := RenderToFile("module {{ .Module }}", dst, map[string]string{"Module": "user"}, WriteOptions{})
	if err != nil {
		t.Fatalf("RenderToFile returned error: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(got) != "module user" {
		t.Fatalf("unexpected file content: %q", got)
	}
}

func TestWriteFileExistingWithoutForceReturnsError(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(dst, []byte("old"), 0o644); err != nil {
		t.Fatalf("WriteFile setup returned error: %v", err)
	}

	err := WriteFile(dst, []byte("new"), WriteOptions{})
	if !errors.Is(err, ErrFileExists) {
		t.Fatalf("expected ErrFileExists, got %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(got) != "old" {
		t.Fatalf("file was overwritten without force: %q", got)
	}
}

func TestWriteFileForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(dst, []byte("old"), 0o644); err != nil {
		t.Fatalf("WriteFile setup returned error: %v", err)
	}

	if err := WriteFile(dst, []byte("new"), WriteOptions{Force: true}); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(got) != "new" {
		t.Fatalf("file was not overwritten: %q", got)
	}
}
