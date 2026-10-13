package preflight

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/huh"
)

type tool struct {
	name       string
	brewPkg    string
	installCmd []string // fallback if brew not available (run with sudo)
	docsURL    string
}

var required = []tool{
	{
		name:    "uv",
		brewPkg: "uv",
		installCmd: []string{
			"sh", "-c", "curl -LsSf https://astral.sh/uv/install.sh | sh",
		},
		docsURL: "https://docs.astral.sh/uv/getting-started/installation/",
	},
	{
		name:    "task",
		brewPkg: "go-task",
		installCmd: []string{
			"sh", "-c", `sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin`,
		},
		docsURL: "https://taskfile.dev/installation/",
	},
}

func Check() error {
	var missing []tool
	for _, t := range required {
		if _, err := exec.LookPath(t.name); err != nil {
			missing = append(missing, t)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	fmt.Println()
	for _, t := range missing {
		fmt.Printf("  ✗ %s is not installed\n", t.name)
	}
	fmt.Println()

	hasBrew := false
	if runtime.GOOS == "darwin" {
		if _, err := exec.LookPath("brew"); err == nil {
			hasBrew = true
		}
	}

	// Non-brew installs write to system paths and require sudo.
	confirmTitle := "Required tools are missing. Install them now?"
	if !hasBrew {
		fmt.Println("  ⚠️  Installing these tools requires sudo access.")
		fmt.Println()
		confirmTitle = "Required tools are missing. Install them now? (requires sudo)"
	}

	var install bool
	prompt := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(confirmTitle).
				Affirmative("Yes, install").
				Negative("No, exit").
				Value(&install),
		),
	)
	if err := prompt.Run(); err != nil {
		return err
	}

	if !install {
		fmt.Println("\nTo install manually, visit:")
		for _, t := range missing {
			fmt.Printf("  • %s: %s\n", t.name, t.docsURL)
		}
		fmt.Println()
		return fmt.Errorf("required tools not installed")
	}

	for _, t := range missing {
		if err := installTool(t, hasBrew); err != nil {
			return fmt.Errorf("failed to install %s: %w", t.name, err)
		}
	}

	// Re-check after install
	for _, t := range missing {
		if _, err := exec.LookPath(t.name); err != nil {
			// Tool may have been installed to a path not yet in this process's PATH.
			// Try common locations before giving up.
			if !existsInCommonPaths(t.name) {
				return fmt.Errorf("%s was installed but is not on PATH — open a new terminal and try again", t.name)
			}
		}
	}

	return nil
}

func installTool(t tool, hasBrew bool) error {
	fmt.Printf("\n  Installing %s...\n", t.name)

	var cmd *exec.Cmd
	if hasBrew {
		cmd = exec.Command("brew", "install", t.brewPkg)
	} else {
		// Non-brew installs write to system paths (e.g. /usr/local/bin) and require sudo.
		sudoArgs := append([]string{"sudo"}, t.installCmd...)
		cmd = exec.Command(sudoArgs[0], sudoArgs[1:]...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func existsInCommonPaths(name string) bool {
	common := []string{
		"/usr/local/bin/" + name,
		"/opt/homebrew/bin/" + name,
		os.ExpandEnv("$HOME/.cargo/bin/" + name),
		os.ExpandEnv("$HOME/.local/bin/" + name),
	}
	for _, p := range common {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}
