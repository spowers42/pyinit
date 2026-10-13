// Package preflight ensures the external tools pyinit depends on are installed.
package preflight

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/huh"

	"github.com/spowers42/pyinit/internal/runner"
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

// Env is the seam over everything Check touches outside its own decisions.
type Env interface {
	// LookPath reports whether name is on PATH.
	LookPath(name string) bool
	// OS returns the operating system identifier (runtime.GOOS).
	OS() string
	// Confirm asks the user a yes/no question.
	Confirm(title string) (bool, error)
	// FileExists reports whether path exists on disk.
	FileExists(path string) bool
	// Runner executes install commands.
	Runner() runner.Runner
	// Out is where user-facing messages are written.
	Out() io.Writer
}

// RealEnv is the production Env.
type RealEnv struct{}

// LookPath reports whether name is on PATH.
func (RealEnv) LookPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// OS returns runtime.GOOS.
func (RealEnv) OS() string { return runtime.GOOS }

// Runner returns a runner.Exec.
func (RealEnv) Runner() runner.Runner { return runner.Exec{} }

// Out returns os.Stdout.
func (RealEnv) Out() io.Writer { return os.Stdout }

// FileExists reports whether path exists.
func (RealEnv) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Confirm shows an install yes/no prompt.
func (RealEnv) Confirm(title string) (bool, error) {
	var ok bool
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Affirmative("Yes, install").
				Negative("No, exit").
				Value(&ok),
		),
	).Run()
	return ok, err
}

// Check finds missing required tools and, with the user's consent, installs them.
func Check(env Env) error {
	out := env.Out()
	var missing []tool
	for _, t := range required {
		if !env.LookPath(t.name) {
			missing = append(missing, t)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	say(out, "\n")
	for _, t := range missing {
		say(out, "  ✗ %s is not installed\n", t.name)
	}
	say(out, "\n")

	hasBrew := env.OS() == "darwin" && env.LookPath("brew")

	// Non-brew installs write to system paths and require sudo.
	confirmTitle := "Required tools are missing. Install them now?"
	if !hasBrew {
		say(out, "  ⚠️  Installing these tools requires sudo access.\n")
		say(out, "\n")
		confirmTitle = "Required tools are missing. Install them now? (requires sudo)"
	}

	install, err := env.Confirm(confirmTitle)
	if err != nil {
		return err
	}

	if !install {
		say(out, "\nTo install manually, visit:\n")
		for _, t := range missing {
			say(out, "  • %s: %s\n", t.name, t.docsURL)
		}
		say(out, "\n")
		return fmt.Errorf("required tools not installed")
	}

	for _, t := range missing {
		if err := installTool(env, t, hasBrew); err != nil {
			return fmt.Errorf("failed to install %s: %w", t.name, err)
		}
	}

	// Re-check after install
	for _, t := range missing {
		if !env.LookPath(t.name) {
			// Tool may have been installed to a path not yet in this process's PATH.
			// Try common locations before giving up.
			if !existsInCommonPaths(env, t.name) {
				return fmt.Errorf("%s was installed but is not on PATH — open a new terminal and try again", t.name)
			}
		}
	}

	return nil
}

func installTool(env Env, t tool, hasBrew bool) error {
	say(env.Out(), "\n  Installing %s...\n", t.name)

	if hasBrew {
		return env.Runner().Run("", "brew", "install", t.brewPkg)
	}
	// Non-brew installs write to system paths (e.g. /usr/local/bin) and require sudo.
	return env.Runner().Run("", "sudo", t.installCmd...)
}

// say writes a user-facing message. Write errors to the terminal are not actionable, so they are dropped.
func say(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprintf(w, format, a...)
}

func existsInCommonPaths(env Env, name string) bool {
	common := []string{
		"/usr/local/bin/" + name,
		"/opt/homebrew/bin/" + name,
		os.ExpandEnv("$HOME/.cargo/bin/" + name),
		os.ExpandEnv("$HOME/.local/bin/" + name),
	}
	for _, p := range common {
		if env.FileExists(p) {
			return true
		}
	}
	return false
}
