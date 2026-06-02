package scaffold

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/scottp/pyinit/internal/tui"
)

//go:embed templates/*
var templateFS embed.FS

func Run(cfg tui.ProjectConfig) error {
	projectDir := filepath.Join(cfg.OutputDir, cfg.ProjectName)

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("create project dir: %w", err)
	}

	fmt.Printf("\n  Running uv init...\n")
	uvInit := exec.Command("uv", "init", "--python", cfg.PythonVersion, ".")
	uvInit.Dir = projectDir
	uvInit.Stdout = os.Stdout
	uvInit.Stderr = os.Stderr
	if err := uvInit.Run(); err != nil {
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

	if err := renderTemplate(cfg, "templates/pyproject.toml.tmpl", filepath.Join(projectDir, "pyproject.toml")); err != nil {
		return err
	}
	if err := renderTemplate(cfg, "templates/README.md.tmpl", filepath.Join(projectDir, "README.md")); err != nil {
		return err
	}
	if err := renderTemplate(cfg, "templates/Taskfile.yml.tmpl", filepath.Join(projectDir, "Taskfile.yml")); err != nil {
		return err
	}
	if err := renderTemplate(cfg, "templates/gitignore.tmpl", filepath.Join(projectDir, ".gitignore")); err != nil {
		return err
	}

	fmt.Printf("  Running uv venv...\n")
	uvVenv := exec.Command("uv", "venv")
	uvVenv.Dir = projectDir
	uvVenv.Stdout = os.Stdout
	uvVenv.Stderr = os.Stderr
	if err := uvVenv.Run(); err != nil {
		return fmt.Errorf("uv venv: %w", err)
	}

	return nil
}

func renderTemplate(cfg tui.ProjectConfig, tmplPath, dstPath string) error {
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
