package tui

import (
	"os"

	"github.com/charmbracelet/huh"

	"github.com/scottp/pyinit/internal/config"
)

func Run() (config.ProjectConfig, error) {
	home, _ := os.UserHomeDir()

	var cfg config.Input
	cfg.PythonVersion = "3.12"
	cfg.OutputDir = home

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Description("lowercase, hyphens and underscores allowed").
				Placeholder("my-project").
				Validate(config.ValidateName).
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
				Validate(config.ValidatePythonVersion).
				Value(&cfg.PythonVersion),

			huh.NewInput().
				Title("Output directory").
				Description("Parent directory where the project folder will be created").
				Placeholder(home).
				Value(&cfg.OutputDir),
		),
	)

	if err := form.Run(); err != nil {
		return config.ProjectConfig{}, err
	}

	return config.New(cfg)
}
