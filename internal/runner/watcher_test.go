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

func TestWatcherSkipsGeneratedAndRuntimeDirectories(t *testing.T) {
	dir := t.TempDir()
	for _, path := range []string{
		"logs/app.log",
		"storage/data.json",
		"dist/assets/index.js",
		"coverage/lcov.info",
		".vite/deps/index.js",
		"node_modules/vite/index.js",
		"runtime/teval/runs/1/output/manifest.json",
		"cache/state.json",
		"temp/probe.json",
		"run/worker.json",
		".output/server/manifest.json",
		".turbo/state.json",
		".parcel-cache/state.json",
		"out/export/manifest.json",
		"target/debug/build.json",
	} {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("MkdirAll returned error: %v", err)
		}
		if err := os.WriteFile(full, []byte("skip\n"), 0o644); err != nil {
			t.Fatalf("WriteFile returned error: %v", err)
		}
	}

	watcher := NewWatcher(dir, []string{"."})
	snapshot, err := watcher.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot returned error: %v", err)
	}
	if len(snapshot) != 0 {
		t.Fatalf("expected runtime/generated files to be skipped, got %v", snapshot)
	}
}

func TestWatcherSkipsGoTestFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app", "server", "main_test.go")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(path, []byte("package server\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	watcher := NewWatcher(dir, []string{"app"})
	snapshot, err := watcher.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot returned error: %v", err)
	}
	if len(snapshot) != 0 {
		t.Fatalf("expected test files to be skipped, got %v", snapshot)
	}
}
