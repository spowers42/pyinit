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
	"slices"

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

// systemDirs are the system-wide install directories Locate checks after PATH.
// A var so tests can substitute writable directories.
var systemDirs = []string{"/usr/local/bin", "/opt/homebrew/bin"}

// Locate looks for name on PATH, then in common install directories that may not be on PATH.
func (RealEnv) Locate(name string) (Location, bool) {
	if p, err := exec.LookPath(name); err == nil {
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}
		return Location{Path: p, OnPath: true}, true
	}
	home := os.Getenv("HOME")
	dirs := append(slices.Clone(systemDirs),
		filepath.Join(home, ".cargo", "bin"),
		filepath.Join(home, ".local", "bin"),
	)
	// uv's installer honours $XDG_BIN_HOME.
	if xdg := os.Getenv("XDG_BIN_HOME"); xdg != "" {
		dirs = append(dirs, xdg)
	}
	for _, dir := range dirs {
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

	var brewPath string
	if brew, found := env.Locate("brew"); found && env.OS() == "darwin" {
		brewPath = brew.Path
	}

	confirmTitle := installPrompt
	if slices.ContainsFunc(missing, func(t tool) bool { return t.needsSudo(brewPath) }) {
		say(out, "  ⚠️  Installing these tools requires sudo access.\n")
		say(out, "\n")
		confirmTitle = installPromptSudo
	}

	confirmed, err := env.Confirm(confirmTitle)
	if err != nil {
		return err
	}

	if !confirmed {
		say(out, "\nTo install manually, visit:\n")
		for _, t := range missing {
			say(out, "  • %s: %s\n", t.name, t.docsURL)
		}
		say(out, "\n")
		return ErrNotInstalled
	}

	for _, t := range missing {
		say(out, "\n  Installing %s...\n", t.name)
		cmd := t.installCommand(brewPath)
		if err := env.Runner().Run("", cmd[0], cmd[1:]...); err != nil {
			return fmt.Errorf("failed to install %s: %w", t.name, err)
		}
	}
	return nil
}

// needsSudo reports whether installing t needs sudo: only its fallback script
// can, and only when brew (at brewPath) is unavailable.
func (t tool) needsSudo(brewPath string) bool {
	return brewPath == "" && t.sudo
}

// installCommand returns the command line that installs t: brew from brewPath
// if set, otherwise t's fallback script, under sudo if t.needsSudo.
func (t tool) installCommand(brewPath string) []string {
	if brewPath != "" {
		return []string{brewPath, "install", t.brewPkg}
	}
	if t.needsSudo(brewPath) {
		return append([]string{"sudo"}, t.installCmd...)
	}
	return t.installCmd
}

// say writes a user-facing message. Write errors to the terminal are not actionable, so they are dropped.
func say(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprintf(w, format, a...)
}
