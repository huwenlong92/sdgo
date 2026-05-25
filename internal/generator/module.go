package generator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/huwenlong/sdgo/internal/project"
	"github.com/huwenlong/sdgo/internal/renderer"
	"github.com/huwenlong/sdgo/internal/templatefs"
)

type ModuleOptions struct {
	Name     string
	WithDocs bool
	Force    bool
}

func GenerateModule(projectDir string, opt ModuleOptions) error {
	if !project.IsGoProject(projectDir) {
		return fmt.Errorf("current directory is not a Go project: go.mod not found")
	}

	data, err := normalizeName(opt.Name)
	if err != nil {
		return err
	}
	data.Year = time.Now().Year()

	opt.WithDocs = true
	dst := filepath.Join(projectDir, "modules", data.SnakeName)
	writeOpt := renderer.WriteOptions{Force: opt.Force}
	if err := renderer.CopyTemplateDir(templatefs.Templates, "templates/module/basic/code", dst, data, writeOpt); err != nil {
		return err
	}
	if err := renderModuleDoc(projectDir, "templates/module/basic/docs/modules/module.md.tmpl", filepath.Join("docs", "modules", data.SnakeName+".md"), data, writeOpt); err != nil {
		return err
	}
	return renderModuleDoc(projectDir, "templates/module/basic/docs/usage/module.md.tmpl", filepath.Join("docs", "usage", data.SnakeName+".md"), data, writeOpt)
}

func renderModuleDoc(projectDir string, tplPath string, dst string, data TemplateData, opt renderer.WriteOptions) error {
	content, err := templatefs.Templates.ReadFile(tplPath)
	if err != nil {
		return fmt.Errorf("template not found: %s: %w", tplPath, err)
	}
	if err := renderer.RenderToFile(string(content), filepath.Join(projectDir, dst), data, opt); err != nil {
		if errors.Is(err, renderer.ErrFileExists) {
			fmt.Fprintf(os.Stderr, "document already exists, skipped: %s\n", dst)
			return nil
		}
		return err
	}
	return nil
}
