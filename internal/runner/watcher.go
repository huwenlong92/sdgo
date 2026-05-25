package runner

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Watcher struct {
	dir   string
	roots []string
}

type FileState struct {
	ModTime time.Time
	Size    int64
}

type Snapshot map[string]FileState

func NewWatcher(dir string, roots []string) *Watcher {
	return &Watcher{dir: dir, roots: roots}
}

func (w *Watcher) Snapshot() (Snapshot, error) {
	out := Snapshot{}
	for _, root := range w.roots {
		path := root
		if !filepath.IsAbs(path) {
			path = filepath.Join(w.dir, root)
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		if err := filepath.WalkDir(path, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if shouldSkipDir(entry.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			if !shouldWatchFile(path) {
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(w.dir, path)
			if err != nil {
				return err
			}
			out[rel] = FileState{ModTime: info.ModTime(), Size: info.Size()}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s Snapshot) Equal(other Snapshot) bool {
	if len(s) != len(other) {
		return false
	}
	for path, state := range s {
		next, ok := other[path]
		if !ok {
			return false
		}
		if !state.ModTime.Equal(next.ModTime) || state.Size != next.Size {
			return false
		}
	}
	return true
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".git", ".gitnexus", ".cache", ".claude", ".codex", ".idea", ".next", ".nuxt", ".svelte-kit", ".tmp", ".vite", ".vscode", "bin", "build", "coverage", "dist", "logs", "node_modules", "storage", "tmp", "vendor":
		return true
	default:
		return false
	}
}

func shouldWatchFile(path string) bool {
	if strings.HasSuffix(strings.ToLower(filepath.Base(path)), "_test.go") {
		return false
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go", ".tpl", ".tmpl", ".html", ".yaml", ".yml", ".json":
		return true
	default:
		return false
	}
}
