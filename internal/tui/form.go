package tui

import (
	"os"

	"github.com/charmbracelet/huh"

	"github.com/scottp/pyinit/internal/config"
)

func Run() (config.ProjectConfig, error) {
	home, _ := os.UserHomeDir()

	var input config.Input
	input.PythonVersion = "3.12"
	input.OutputDir = home

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Description("lowercase, hyphens and underscores allowed").
				Placeholder("my-project").
				Validate(config.ValidateName).
				Value(&input.ProjectName),

			huh.NewInput().
				Title("Description").
				Placeholder("A short description of your project").
				Value(&input.Description),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Author name").
				Placeholder("Jane Smith").
				Value(&input.AuthorName),

			huh.NewInput().
				Title("Author email").
				Placeholder("jane@example.com").
				Value(&input.AuthorEmail),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Python version").
				Placeholder("3.12").
				Validate(config.ValidatePythonVersion).
				Value(&input.PythonVersion),

			huh.NewInput().
				Title("Output directory").
				Description("Parent directory where the project folder will be created").
				Placeholder(home).
				Validate(config.ValidateOutputDir).
				Value(&input.OutputDir),
		),
	)

	if err := form.Run(); err != nil {
		return config.ProjectConfig{}, err
	}

	return config.New(input)
}
