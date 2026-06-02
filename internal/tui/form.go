package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/charmbracelet/huh"
)

type ProjectConfig struct {
	ProjectName      string
	PackageName      string
	Description      string
	AuthorName       string
	AuthorEmail      string
	PythonVersion    string
	PythonVersionTag string // e.g. "312" for 3.12
	OutputDir        string
}

var validName = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

func Run() (ProjectConfig, error) {
	home, _ := os.UserHomeDir()

	var cfg ProjectConfig
	cfg.PythonVersion = "3.12"
	cfg.OutputDir = home

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Description("lowercase, hyphens and underscores allowed").
				Placeholder("my-project").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("project name is required")
					}
					if !validName.MatchString(s) {
						return fmt.Errorf("must be lowercase and start with a letter (a-z, 0-9, - and _ allowed)")
					}
					return nil
				}).
				Value(&cfg.ProjectName),

			huh.NewInput().
				Title("Description").
				Placeholder("A short description of your project").
				Value(&cfg.Description),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Author name").
				Placeholder("Jane Smith").
				Value(&cfg.AuthorName),

			huh.NewInput().
				Title("Author email").
				Placeholder("jane@example.com").
				Value(&cfg.AuthorEmail),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Python version").
				Placeholder("3.12").
				Value(&cfg.PythonVersion),

			huh.NewInput().
				Title("Output directory").
				Description("Parent directory where the project folder will be created").
				Placeholder(home).
				Value(&cfg.OutputDir),
		),
	)

	if err := form.Run(); err != nil {
		return cfg, err
	}

	if cfg.OutputDir == "" {
		cfg.OutputDir = home
	}

	expanded, err := expandPath(cfg.OutputDir)
	if err != nil {
		return cfg, err
	}
	cfg.OutputDir = expanded

	cfg.PackageName = toPackageName(cfg.ProjectName)
	cfg.PythonVersionTag = toPythonVersionTag(cfg.PythonVersion)

	return cfg, nil
}

func toPackageName(name string) string {
	result := make([]byte, len(name))
	for i := range name {
		if name[i] == '-' {
			result[i] = '_'
		} else {
			result[i] = name[i]
		}
	}
	return string(result)
}

// toPythonVersionTag converts "3.12" → "312" for use in ruff's target-version.
func toPythonVersionTag(version string) string {
	out := make([]byte, 0, len(version))
	for i := range version {
		if version[i] != '.' {
			out = append(out, version[i])
		}
	}
	return string(out)
}

func expandPath(path string) (string, error) {
	if len(path) == 0 {
		return path, nil
	}
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[1:]), nil
	}
	return filepath.Abs(path)
}
