package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/scottp/pyinit/internal/preflight"
	"github.com/scottp/pyinit/internal/scaffold"
	"github.com/scottp/pyinit/internal/tui"
)

func main() {
	fmt.Println("pyinit — bootstrap a new Python project")
	fmt.Println()

	if err := preflight.Check(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	cfg, err := tui.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nCreating project %q in %s...\n", cfg.ProjectName, cfg.OutputDir)

	if err := scaffold.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	projectDir := filepath.Join(cfg.OutputDir, cfg.ProjectName)
	fmt.Printf("\nDone! Your project is ready at %s\n\n", projectDir)
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", projectDir)
	fmt.Println("  task install    # install dependencies")
	fmt.Println("  task check      # lint + test")
}
