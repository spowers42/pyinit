// Package preflight ensures the external tools pyinit depends on are installed.
package preflight

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/charmbracelet/huh"

	"github.com/spowers42/pyinit/internal/runner"
)

// ErrNotInstalled is returned when the user declines to install missing tools.
var ErrNotInstalled = errors.New("required tools not installed")

// Install prompt titles; the sudo variant is shown when an install needs sudo.
const (
	installPrompt     = "Required tools are missing. Install them now?"
	installPromptSudo = installPrompt + " (requires sudo)"
)

type tool struct {
	name       string
	brewPkg    string
	installCmd []string // fallback if brew not available
	sudo       bool     // whether installCmd must run under sudo
	docsURL    string
}

var required = []tool{
	{
		name:    "uv",
		brewPkg: "uv",
		// uv's installer is per-user ($XDG_BIN_HOME, else ~/.local/bin); under sudo it would land in root's home.
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
		sudo:    true, // installs to /usr/local/bin
		docsURL: "https://taskfile.dev/installation/",
	},
}

// ToolNames returns the names of every required tool.
func ToolNames() []string {
	names := make([]string, len(required))
	for i, t := range required {
		names[i] = t.name
	}
	return names
}

// Location is where a tool was found, and whether that place is on the user's PATH.
type Location struct {
	Path   string
	OnPath bool
}

// Env is the seam over everything Check touches outside its own decisions.
type Env interface {
	// Locate reports where the tool name is, if anywhere.
	Locate(name string) (Location, bool)
	// OS returns the operating system identifier (runtime.GOOS).
	OS() string
	// Confirm asks the user a yes/no question.
	Confirm(title string) (bool, error)
	// Runner executes install commands.
	Runner() runner.Runner
	// Out is where user-facing messages are written.
	Out() io.Writer
}

// RealEnv is the production Env.
type RealEnv struct{}

// Locate looks for name on PATH, then in common install directories that may not be on PATH.
func (RealEnv) Locate(name string) (Location, bool) {
	if p, err := exec.LookPath(name); err == nil {
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}
		return Location{Path: p, OnPath: true}, true
	}
	home := os.Getenv("HOME")
	for _, dir := range []string{
		"/usr/local/bin",
		"/opt/homebrew/bin",
		filepath.Join(home, ".cargo", "bin"),
		filepath.Join(home, ".local", "bin"),
	} {
		p := filepath.Join(dir, name)
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return Location{Path: p, OnPath: false}, true
		}
	}
	return Location{}, false
}

// OS returns runtime.GOOS.
func (RealEnv) OS() string { return runtime.GOOS }

// Runner returns a runner.Exec.
func (RealEnv) Runner() runner.Runner { return runner.Exec{} }

// Out returns os.Stdout.
func (RealEnv) Out() io.Writer { return os.Stdout }

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

// Check finds a Location for every required tool, installing missing ones with
// the user's consent. It returns a resolving runner that runs each required tool
// from its Location, wrapping env.Runner().
func Check(env Env) (runner.Runner, error) {
	out := env.Out()
	paths := map[string]string{}
	resolve := func(t tool) bool {
		loc, ok := env.Locate(t.name)
		if !ok {
			return false
		}
		paths[t.name] = loc.Path
		if !loc.OnPath {
			say(out, "  ⚠️  %s was found in %s, which is not on your PATH. Add it to PATH so the generated project's tasks can find it.\n",
				t.name, filepath.Dir(loc.Path))
		}
		return true
	}

	var missing []tool
	for _, t := range required {
		if !resolve(t) {
			missing = append(missing, t)
		}
	}

	if len(missing) > 0 {
		if err := install(env, missing); err != nil {
			return nil, err
		}
		for _, t := range missing {
			if !resolve(t) {
				return nil, fmt.Errorf("%s was installed but could not be found — open a new terminal and try again", t.name)
			}
		}
	}

	return runner.Resolving(env.Runner(), paths), nil
}

// install asks for consent and installs the missing tools.
func install(env Env, missing []tool) error {
	out := env.Out()
	say(out, "\n")
	for _, t := range missing {
		say(out, "  ✗ %s is not installed\n", t.name)
	}
	say(out, "\n")

	brew, ok := env.Locate("brew")
	hasBrew := env.OS() == "darwin" && ok && brew.OnPath

	needsSudo := false
	if !hasBrew {
		for _, t := range missing {
			needsSudo = needsSudo || t.sudo
		}
	}

	confirmTitle := installPrompt
	if needsSudo {
		say(out, "  ⚠️  Installing these tools requires sudo access.\n")
		say(out, "\n")
		confirmTitle = installPromptSudo
	}

	ok, err := env.Confirm(confirmTitle)
	if err != nil {
		return err
	}

	if !ok {
		say(out, "\nTo install manually, visit:\n")
		for _, t := range missing {
			say(out, "  • %s: %s\n", t.name, t.docsURL)
		}
		say(out, "\n")
		return ErrNotInstalled
	}

	for _, t := range missing {
		if err := installTool(env, t, hasBrew); err != nil {
			return fmt.Errorf("failed to install %s: %w", t.name, err)
		}
	}
	return nil
}

func installTool(env Env, t tool, hasBrew bool) error {
	say(env.Out(), "\n  Installing %s...\n", t.name)

	if hasBrew {
		return env.Runner().Run("", "brew", "install", t.brewPkg)
	}
	if t.sudo {
		return env.Runner().Run("", "sudo", t.installCmd...)
	}
	return env.Runner().Run("", t.installCmd[0], t.installCmd[1:]...)
}

// say writes a user-facing message. Write errors to the terminal are not actionable, so they are dropped.
func say(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprintf(w, format, a...)
}
