package preflight

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/spowers42/pyinit/internal/runner"
)

const (
	sudoPrompt  = "Required tools are missing. Install them now? (requires sudo)"
	plainPrompt = "Required tools are missing. Install them now?"
)

type fakeEnv struct {
	onPath  map[string]bool
	goos    string
	confirm bool
	files   map[string]bool
	run     runner.Fake
	out     bytes.Buffer
	prompts []string
}

func (f *fakeEnv) LookPath(n string) bool   { return f.onPath[n] }
func (f *fakeEnv) OS() string               { return f.goos }
func (f *fakeEnv) FileExists(p string) bool { return f.files[p] }
func (f *fakeEnv) Runner() runner.Runner    { return &f.run }
func (f *fakeEnv) Out() io.Writer           { return &f.out }
func (f *fakeEnv) Confirm(t string) (bool, error) {
	f.prompts = append(f.prompts, t)
	return f.confirm, nil
}

func newEnv(goos string, onPath ...string) *fakeEnv {
	e := &fakeEnv{goos: goos, onPath: map[string]bool{}, files: map[string]bool{}, confirm: true}
	for _, p := range onPath {
		e.onPath[p] = true
	}
	return e
}

// installsOntoPath makes every install command put tools on PATH, as a successful install would.
func (f *fakeEnv) installsOntoPath() {
	f.run.OnRun = func(runner.Call) error {
		for _, t := range required {
			f.onPath[t.name] = true
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

func TestCheckAllPresent(t *testing.T) {
	e := newEnv("linux", "uv", "task")
	if err := Check(e); err != nil {
		t.Fatal(err)
	}
	assertPrompts(t, e)
	assertCalls(t, e)
	if e.out.Len() != 0 {
		t.Fatalf("unexpected output: %q", e.out.String())
	}
}

func TestCheckDecline(t *testing.T) {
	e := newEnv("linux")
	e.confirm = false
	err := Check(e)
	if err == nil || err.Error() != "required tools not installed" {
		t.Fatalf("err = %v", err)
	}
	assertPrompts(t, e, sudoPrompt)
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
	if err := Check(e); err != nil {
		t.Fatal(err)
	}
	assertPrompts(t, e, plainPrompt)
	assertCalls(t, e, "brew install uv", "brew install go-task")
	if strings.Contains(e.out.String(), "sudo") {
		t.Fatal("brew path must not mention sudo")
	}
}

func TestCheckSudoInstall(t *testing.T) {
	e := newEnv("linux")
	e.installsOntoPath()
	if err := Check(e); err != nil {
		t.Fatal(err)
	}
	assertPrompts(t, e, sudoPrompt)
	assertCalls(t, e,
		"sudo sh -c curl -LsSf https://astral.sh/uv/install.sh | sh",
		`sudo sh -c sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin`,
	)
	if !strings.Contains(e.out.String(), "Installing these tools requires sudo access.") {
		t.Fatalf("sudo warning missing from %q", e.out.String())
	}
}

func TestCheckDarwinWithoutBrewUsesSudo(t *testing.T) {
	e := newEnv("darwin", "task")
	e.installsOntoPath()
	if err := Check(e); err != nil {
		t.Fatal(err)
	}
	assertPrompts(t, e, sudoPrompt)
	assertCalls(t, e, "sudo sh -c curl -LsSf https://astral.sh/uv/install.sh | sh")
}

func TestCheckInstallFailureWrapped(t *testing.T) {
	e := newEnv("linux", "task")
	boom := errors.New("boom")
	e.run.OnRun = func(runner.Call) error { return boom }
	err := Check(e)
	if !errors.Is(err, boom) || !strings.Contains(err.Error(), "failed to install uv") {
		t.Fatalf("err = %v", err)
	}
}

func TestCheckRecheckNotOnPath(t *testing.T) {
	e := newEnv("linux", "task")
	err := Check(e)
	if err == nil || !strings.Contains(err.Error(), "uv was installed but is not on PATH") {
		t.Fatalf("err = %v", err)
	}
}

func TestCheckRecheckCommonPathsFallback(t *testing.T) {
	e := newEnv("linux", "task")
	e.files["/usr/local/bin/uv"] = true
	if err := Check(e); err != nil {
		t.Fatal(err)
	}
}
