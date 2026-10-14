package preflight

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/spowers42/pyinit/internal/runner"
)

type fakeEnv struct {
	located map[string]Location
	goos    string
	confirm bool
	run     runner.Fake
	out     bytes.Buffer
	prompts []string
}

func (f *fakeEnv) Locate(name string) (Location, bool) {
	loc, ok := f.located[name]
	return loc, ok
}
func (f *fakeEnv) OS() string            { return f.goos }
func (f *fakeEnv) Runner() runner.Runner { return &f.run }
func (f *fakeEnv) Out() io.Writer        { return &f.out }
func (f *fakeEnv) Confirm(title string) (bool, error) {
	f.prompts = append(f.prompts, title)
	return f.confirm, nil
}

// newEnv returns a fake env where each named tool is on PATH in /usr/bin.
func newEnv(goos string, onPath ...string) *fakeEnv {
	e := &fakeEnv{goos: goos, located: map[string]Location{}, confirm: true}
	for _, p := range onPath {
		e.located[p] = Location{Path: "/usr/bin/" + p, OnPath: true}
	}
	return e
}

// installsOntoPath makes every install command put tools on PATH, as a successful install would.
func (f *fakeEnv) installsOntoPath() {
	f.run.OnRun = func(runner.Call) error {
		for _, t := range required {
			f.located[t.name] = Location{Path: "/usr/bin/" + t.name, OnPath: true}
		}
		return nil
	}
}

func assertPrompts(t *testing.T, e *fakeEnv, want ...string) {
	t.Helper()
	if strings.Join(e.prompts, "|") != strings.Join(want, "|") {
		t.Fatalf("prompts = %q, want %q", e.prompts, want)
	}
}

func assertCalls(t *testing.T, e *fakeEnv, want ...string) {
	t.Helper()
	var got []string
	for _, c := range e.run.Calls {
		got = append(got, c.String())
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("calls = %q, want %q", got, want)
	}
}

func assertOutputContains(t *testing.T, e *fakeEnv, want ...string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(e.out.String(), w) {
			t.Errorf("output missing %q: %q", w, e.out.String())
		}
	}
}

// assertResolves runs name through r and checks the env's Fake saw it as path.
func assertResolves(t *testing.T, e *fakeEnv, r runner.Runner, name, path string) {
	t.Helper()
	if r == nil {
		t.Fatal("Check returned nil runner")
	}
	before := len(e.run.Calls)
	if err := r.Run("/proj", name, "--version"); err != nil {
		t.Fatal(err)
	}
	if len(e.run.Calls) != before+1 {
		t.Fatalf("runner made %d calls, want 1", len(e.run.Calls)-before)
	}
	if got := e.run.Calls[before].Name; got != path {
		t.Fatalf("%s resolved to %q, want %q", name, got, path)
	}
}

func TestCheckAllPresent(t *testing.T) {
	e := newEnv("linux", "uv", "task")
	r, err := Check(e)
	if err != nil {
		t.Fatal(err)
	}
	assertPrompts(t, e)
	assertCalls(t, e)
	if e.out.Len() != 0 {
		t.Fatalf("unexpected output: %q", e.out.String())
	}
	assertResolves(t, e, r, "uv", "/usr/bin/uv")
	assertResolves(t, e, r, "task", "/usr/bin/task")
}

func TestCheckOffPathToolSucceedsWithWarning(t *testing.T) {
	e := newEnv("linux", "task")
	e.located["uv"] = Location{Path: "/home/u/.local/bin/uv", OnPath: false}
	r, err := Check(e)
	if err != nil {
		t.Fatal(err)
	}
	// Already installed off PATH: no install prompt, nothing run.
	assertPrompts(t, e)
	assertCalls(t, e)
	if !strings.Contains(e.out.String(), "/home/u/.local/bin") || !strings.Contains(e.out.String(), "PATH") {
		t.Fatalf("warning naming /home/u/.local/bin missing from %q", e.out.String())
	}
	assertResolves(t, e, r, "uv", "/home/u/.local/bin/uv")
}

func TestCheckDecline(t *testing.T) {
	e := newEnv("linux")
	e.confirm = false
	_, err := Check(e)
	if !errors.Is(err, ErrNotInstalled) {
		t.Fatalf("err = %v", err)
	}
	assertPrompts(t, e, installPromptSudo)
	assertCalls(t, e)
	for _, link := range []string{
		"uv: https://docs.astral.sh/uv/getting-started/installation/",
		"task: https://taskfile.dev/installation/",
	} {
		if !strings.Contains(e.out.String(), link) {
			t.Errorf("docs link %q missing from %q", link, e.out.String())
		}
	}
}

func TestCheckBrewInstall(t *testing.T) {
	e := newEnv("darwin", "brew")
	e.installsOntoPath()
	r, err := Check(e)
	if err != nil {
		t.Fatal(err)
	}
	assertPrompts(t, e, installPrompt)
	assertCalls(t, e, "/usr/bin/brew install uv", "/usr/bin/brew install go-task")
	if strings.Contains(e.out.String(), "sudo") {
		t.Fatal("brew path must not mention sudo")
	}
	assertResolves(t, e, r, "uv", "/usr/bin/uv")
}

func TestCheckBrewOffPathInstallsViaAbsolutePath(t *testing.T) {
	e := newEnv("darwin", "uv")
	e.located["brew"] = Location{Path: "/opt/homebrew/bin/brew", OnPath: false}
	e.installsOntoPath()
	if _, err := Check(e); err != nil {
		t.Fatal(err)
	}
	assertPrompts(t, e, installPrompt)
	assertCalls(t, e, "/opt/homebrew/bin/brew install go-task")
}

func TestCheckScriptInstallSudoOnlyForTask(t *testing.T) {
	e := newEnv("linux")
	e.installsOntoPath()
	r, err := Check(e)
	if err != nil {
		t.Fatal(err)
	}
	assertPrompts(t, e, installPromptSudo)
	assertCalls(t, e,
		"sh -c curl -LsSf https://astral.sh/uv/install.sh | sh",
		`sudo sh -c sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin`,
	)
	if !strings.Contains(e.out.String(), "Installing these tools requires sudo access.") {
		t.Fatalf("sudo warning missing from %q", e.out.String())
	}
	assertResolves(t, e, r, "task", "/usr/bin/task")
}

func TestCheckOnlyUVMissingLinuxNoSudo(t *testing.T) {
	e := newEnv("linux", "task")
	e.installsOntoPath()
	if _, err := Check(e); err != nil {
		t.Fatal(err)
	}
	assertPrompts(t, e, installPrompt)
	assertCalls(t, e, "sh -c curl -LsSf https://astral.sh/uv/install.sh | sh")
	if strings.Contains(e.out.String(), "sudo") {
		t.Fatalf("uv-only install must not mention sudo: %q", e.out.String())
	}
}

func TestCheckOnlyTaskMissingLinuxUsesSudo(t *testing.T) {
	e := newEnv("linux", "uv")
	e.installsOntoPath()
	if _, err := Check(e); err != nil {
		t.Fatal(err)
	}
	assertPrompts(t, e, installPromptSudo)
	assertCalls(t, e, `sudo sh -c sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin`)
	assertOutputContains(t, e, "Installing these tools requires sudo access.")
}

func TestCheckTaskMissingDarwinWithoutBrewUsesSudo(t *testing.T) {
	e := newEnv("darwin", "uv")
	e.installsOntoPath()
	if _, err := Check(e); err != nil {
		t.Fatal(err)
	}
	assertPrompts(t, e, installPromptSudo)
	assertCalls(t, e, `sudo sh -c sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin`)
	if !strings.Contains(e.out.String(), "Installing these tools requires sudo access.") {
		t.Fatalf("sudo warning missing from %q", e.out.String())
	}
}

func TestCheckInstallFailureWrapped(t *testing.T) {
	e := newEnv("linux", "task")
	boom := errors.New("boom")
	e.run.OnRun = func(runner.Call) error { return boom }
	_, err := Check(e)
	if !errors.Is(err, boom) || !strings.Contains(err.Error(), "failed to install uv") {
		t.Fatalf("err = %v", err)
	}
}

func TestCheckRecheckNotFound(t *testing.T) {
	e := newEnv("linux", "task")
	_, err := Check(e)
	if err == nil || !strings.Contains(err.Error(), "uv was installed but could not be found") {
		t.Fatalf("err = %v", err)
	}
	assertPrompts(t, e, installPrompt)
	assertCalls(t, e, "sh -c curl -LsSf https://astral.sh/uv/install.sh | sh")
}

func TestCheckRecheckFindsOffPathInstall(t *testing.T) {
	e := newEnv("linux", "task")
	e.run.OnRun = func(runner.Call) error {
		e.located["uv"] = Location{Path: "/home/u/.local/bin/uv", OnPath: false}
		return nil
	}
	r, err := Check(e)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(e.out.String(), "/home/u/.local/bin") {
		t.Fatalf("warning naming /home/u/.local/bin missing from %q", e.out.String())
	}
	assertResolves(t, e, r, "uv", "/home/u/.local/bin/uv")
}
