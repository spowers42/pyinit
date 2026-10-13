package scaffold

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spowers42/pyinit/internal/config"
	"github.com/spowers42/pyinit/internal/runner"
)

//go:embed templates/*
var templateFS embed.FS

// templates maps each embedded template to its destination relative to the project dir.
var templates = []struct{ src, dst string }{
	{"templates/pyproject.toml.tmpl", "pyproject.toml"},
	{"templates/README.md.tmpl", "README.md"},
	{"templates/Taskfile.yml.tmpl", "Taskfile.yml"},
	{"templates/gitignore.tmpl", ".gitignore"},
}

// Run creates the project on disk, executing external commands through r.
func Run(cfg config.ProjectConfig, r runner.Runner) error {
	projectDir := filepath.Join(cfg.OutputDir, cfg.ProjectName)

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("create project dir: %w", err)
	}

	fmt.Printf("\n  Running uv init...\n")
	if err := r.Run(projectDir, "uv", "init", "--python", cfg.PythonVersion, "."); err != nil {
		return fmt.Errorf("uv init: %w", err)
	}

	srcPkg := filepath.Join(projectDir, "src", cfg.PackageName)
	if err := os.MkdirAll(srcPkg, 0755); err != nil {
		return fmt.Errorf("create src dir: %w", err)
	}

	mainSrc := filepath.Join(projectDir, "main.py")
	mainDst := filepath.Join(srcPkg, "main.py")
	if _, err := os.Stat(mainSrc); err == nil {
		if err := os.Rename(mainSrc, mainDst); err != nil {
			return fmt.Errorf("move main.py: %w", err)
		}
	}

	initFile := filepath.Join(srcPkg, "__init__.py")
	if err := os.WriteFile(initFile, []byte(""), 0644); err != nil {
		return fmt.Errorf("create __init__.py: %w", err)
	}

	testsDir := filepath.Join(projectDir, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		return fmt.Errorf("create tests dir: %w", err)
	}
	testsInit := filepath.Join(testsDir, "__init__.py")
	if err := os.WriteFile(testsInit, []byte(""), 0644); err != nil {
		return fmt.Errorf("create tests/__init__.py: %w", err)
	}

	docsDir := filepath.Join(projectDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return fmt.Errorf("create docs dir: %w", err)
	}

	for _, t := range templates {
		if err := renderTemplate(cfg, t.src, filepath.Join(projectDir, t.dst)); err != nil {
			return err
		}
	}

	fmt.Printf("  Running uv venv...\n")
	if err := r.Run(projectDir, "uv", "venv"); err != nil {
		return fmt.Errorf("uv venv: %w", err)
	}

	return nil
}

func renderTemplate(cfg config.ProjectConfig, tmplPath, dstPath string) (err error) {
	data, err := templateFS.ReadFile(tmplPath)
	if err != nil {
		return fmt.Errorf("read template %s: %w", tmplPath, err)
	}

	tmpl, err := template.New(filepath.Base(tmplPath)).Parse(string(data))
	if err != nil {
		return fmt.Errorf("parse template %s: %w", tmplPath, err)
	}

	f, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", dstPath, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	return tmpl.Execute(f, cfg)
}
