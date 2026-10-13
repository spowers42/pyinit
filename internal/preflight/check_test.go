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
	onPath  map[string]bool
	os      string
	confirm bool
	files   map[string]bool
	run     runner.Fake
	out     bytes.Buffer
	prompts []string
}

func (f *fakeEnv) LookPath(n string) bool   { return f.onPath[n] }
func (f *fakeEnv) OS() string               { return f.os }
func (f *fakeEnv) FileExists(p string) bool { return f.files[p] }
func (f *fakeEnv) Runner() runner.Runner    { return &f.run }
func (f *fakeEnv) Out() io.Writer           { return &f.out }
func (f *fakeEnv) Confirm(t string) (bool, error) {
	f.prompts = append(f.prompts, t)
	return f.confirm, nil
}

func newEnv(os string, onPath ...string) *fakeEnv {
	e := &fakeEnv{os: os, onPath: map[string]bool{}, files: map[string]bool{}, confirm: true}
	for _, p := range onPath {
		e.onPath[p] = true
	}
	return e
}

func TestAllPresent(t *testing.T) {
	e := newEnv("linux", "uv", "task")
	if err := Check(e); err != nil {
		t.Fatal(err)
	}
	if len(e.prompts) != 0 || len(e.run.Calls) != 0 || e.out.Len() != 0 {
		t.Fatal("expected no side effects")
	}
}

func TestDecline(t *testing.T) {
	e := newEnv("linux", "task")
	e.confirm = false
	err := Check(e)
	if err == nil || err.Error() != "required tools not installed" {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(e.out.String(), "https://docs.astral.sh/uv/getting-started/installation/") {
		t.Fatalf("docs link missing: %q", e.out.String())
	}
	if len(e.run.Calls) != 0 {
		t.Fatal("must not install")
	}
}

func TestBrewInstall(t *testing.T) {
	e := newEnv("darwin", "brew", "task")
	e.run.OnRun = func(runner.Call) error { e.onPath["uv"] = true; return nil }
	if err := Check(e); err != nil {
		t.Fatal(err)
	}
	if got := e.run.Calls[0].String(); got != "brew install uv" || len(e.run.Calls) != 1 {
		t.Fatalf("calls = %v", e.run.Calls)
	}
	if strings.Contains(e.out.String(), "sudo") || strings.Contains(e.prompts[0], "sudo") {
		t.Fatal("brew path must not mention sudo")
	}
}

func TestSudoInstall(t *testing.T) {
	e := newEnv("linux", "task")
	e.run.OnRun = func(runner.Call) error { e.onPath["uv"] = true; return nil }
	if err := Check(e); err != nil {
		t.Fatal(err)
	}
	if c := e.run.Calls[0]; c.Name != "sudo" || c.Args[0] != "sh" {
		t.Fatalf("call = %v", c)
	}
	if !strings.Contains(e.out.String(), "requires sudo") || !strings.Contains(e.prompts[0], "(requires sudo)") {
		t.Fatal("sudo warning missing")
	}
}

func TestDarwinWithoutBrewUsesSudo(t *testing.T) {
	e := newEnv("darwin", "task")
	e.run.OnRun = func(runner.Call) error { e.onPath["uv"] = true; return nil }
	_ = Check(e)
	if e.run.Calls[0].Name != "sudo" {
		t.Fatalf("call = %v", e.run.Calls[0])
	}
}

func TestInstallFailureWrapped(t *testing.T) {
	e := newEnv("linux", "task")
	boom := errors.New("boom")
	e.run.OnRun = func(runner.Call) error { return boom }
	err := Check(e)
	if !errors.Is(err, boom) || !strings.Contains(err.Error(), "failed to install uv") {
		t.Fatalf("err = %v", err)
	}
}

func TestRecheckNotOnPath(t *testing.T) {
	e := newEnv("linux", "task")
	err := Check(e)
	if err == nil || !strings.Contains(err.Error(), "uv was installed but is not on PATH") {
		t.Fatalf("err = %v", err)
	}
}

func TestRecheckCommonPathsFallback(t *testing.T) {
	e := newEnv("linux", "task")
	e.files["/usr/local/bin/uv"] = true
	if err := Check(e); err != nil {
		t.Fatal(err)
	}
}
