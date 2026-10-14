package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spowers42/pyinit/internal/preflight"
	"github.com/spowers42/pyinit/internal/scaffold"
	"github.com/spowers42/pyinit/internal/tui"
)

// version is set at build time via -ldflags "-X main.version=<tag>".
var version = "dev"

func main() {
	fmt.Printf("pyinit %s — bootstrap a new Python project\n", version)
	fmt.Println()

	run, err := preflight.Check(preflight.RealEnv{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	cfg, err := tui.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nCreating project %q in %s...\n", cfg.ProjectName, cfg.OutputDir)

	if err := scaffold.Run(cfg, run); err != nil {
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
