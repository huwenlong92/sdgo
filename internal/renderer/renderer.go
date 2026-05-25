package renderer

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var ErrFileExists = errors.New("file already exists")

type WriteOptions struct {
	Force bool
}

func Render(tpl string, data any) ([]byte, error) {
	parsed, err := template.New(filepath.Base(tpl)).Parse(tpl)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := parsed.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}
	return buf.Bytes(), nil
}

func WriteFile(path string, data []byte, opt WriteOptions) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}

	if _, err := os.Stat(path); err == nil && !opt.Force {
		return fmt.Errorf("%w: %s, use --force to overwrite", ErrFileExists, path)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat file: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

func RenderToFile(tpl string, dst string, data any, opt WriteOptions) error {
	out, err := Render(tpl, data)
	if err != nil {
		return err
	}
	return WriteFile(dst, out, opt)
}

func CopyTemplateDir(srcFS fs.FS, src string, dst string, data any, opt WriteOptions) error {
	entries, err := fs.ReadDir(srcFS, src)
	if err != nil {
		return fmt.Errorf("template not found: %s: %w", src, err)
	}

	for _, entry := range entries {
		srcPath := filepath.ToSlash(filepath.Join(src, entry.Name()))
		dstName := strings.TrimSuffix(entry.Name(), ".tmpl")
		if strings.HasPrefix(dstName, "DOT_") {
			dstName = "." + strings.TrimPrefix(dstName, "DOT_")
		}
		dstPath := filepath.Join(dst, dstName)

		if entry.IsDir() {
			if err := CopyTemplateDir(srcFS, srcPath, dstPath, data, opt); err != nil {
				return err
			}
			continue
		}

		content, err := fs.ReadFile(srcFS, srcPath)
		if err != nil {
			return fmt.Errorf("read template %s: %w", srcPath, err)
		}
		if strings.HasSuffix(entry.Name(), ".tmpl") {
			if err := RenderToFile(string(content), dstPath, data, opt); err != nil {
				return err
			}
			continue
		}
		if err := WriteFile(dstPath, content, opt); err != nil {
			return err
		}
	}
	return nil
}
