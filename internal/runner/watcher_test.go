package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWatcherDetectsChangedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app", "server", "main.go")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(path, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	watcher := NewWatcher(dir, []string{"app"})
	first, err := watcher.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot returned error: %v", err)
	}

	if err := os.WriteFile(path, []byte("package main\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	second, err := watcher.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot returned error: %v", err)
	}
	if first.Equal(second) {
		t.Fatalf("expected snapshots to differ after file change")
	}
}

func TestWatcherSkipsTmpDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tmp", "main.go")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(path, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	watcher := NewWatcher(dir, []string{"."})
	snapshot, err := watcher.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot returned error: %v", err)
	}
	if len(snapshot) != 0 {
		t.Fatalf("expected tmp files to be skipped, got %v", snapshot)
	}
}
