package scaffold

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spowers42/pyinit/internal/config"
	"github.com/spowers42/pyinit/internal/runner"
)

func testConfig(t *testing.T) config.ProjectConfig {
	t.Helper()
	return config.ProjectConfig{
		ProjectName:      "my-app",
		PackageName:      "my_app",
		Description:      "A test app",
		AuthorName:       "Ada Lovelace",
		AuthorEmail:      "ada@example.com",
		PythonVersion:    "3.12",
		PythonVersionTag: "312",
		OutputDir:        t.TempDir(),
	}
}

// fakeUV mimics uv: `uv init` drops a main.py and a minimal pyproject.toml.
func fakeUV() *runner.Fake {
	return &runner.Fake{OnRun: func(c runner.Call) error {
		if len(c.Args) > 0 && c.Args[0] == "init" {
			if err := os.WriteFile(filepath.Join(c.Dir, "main.py"), []byte("print('hi')\n"), 0644); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(c.Dir, "pyproject.toml"), []byte("[project]\n"), 0644)
		}
		return nil
	}}
}

func TestRunCommandSequence(t *testing.T) {
	cfg := testConfig(t)
	fake := fakeUV()
	if err := Run(cfg, fake); err != nil {
		t.Fatal(err)
	}
	projectDir := filepath.Join(cfg.OutputDir, cfg.ProjectName)
	want := []string{"uv init --python 3.12 .", "uv venv"}
	if len(fake.Calls) != len(want) {
		t.Fatalf("got %d calls, want %d: %v", len(fake.Calls), len(want), fake.Calls)
	}
	for i, w := range want {
		if got := fake.Calls[i].String(); got != w {
			t.Errorf("call %d = %q, want %q", i, got, w)
		}
		if fake.Calls[i].Dir != projectDir {
			t.Errorf("call %d dir = %q, want %q", i, fake.Calls[i].Dir, projectDir)
		}
	}
}

func TestRunLayout(t *testing.T) {
	cfg := testConfig(t)
	if err := Run(cfg, fakeUV()); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(cfg.OutputDir, cfg.ProjectName)

	for _, p := range []string{
		"src/my_app/main.py",
		"src/my_app/__init__.py",
		"tests/__init__.py",
		"docs",
		"pyproject.toml",
		"README.md",
		"Taskfile.yml",
		".gitignore",
	} {
		if _, err := os.Stat(filepath.Join(root, p)); err != nil {
			t.Errorf("missing %s: %v", p, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "main.py")); !os.IsNotExist(err) {
		t.Errorf("root main.py should have been moved, stat err = %v", err)
	}
}

func TestRunRendersTemplates(t *testing.T) {
	cfg := testConfig(t)
	if err := Run(cfg, fakeUV()); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(cfg.OutputDir, cfg.ProjectName)

	pyproject, err := os.ReadFile(filepath.Join(root, "pyproject.toml"))
	if err != nil {
		t.Fatal(err)
	}
	// Rendered template must overwrite uv's minimal pyproject.
	for _, want := range []string{cfg.ProjectName, cfg.PackageName, cfg.AuthorName, cfg.AuthorEmail, cfg.Description} {
		if !strings.Contains(string(pyproject), want) {
			t.Errorf("pyproject.toml missing %q:\n%s", want, pyproject)
		}
	}
	if strings.Contains(string(pyproject), "{{") {
		t.Errorf("pyproject.toml has unrendered template actions:\n%s", pyproject)
	}
}

func TestRunUVFailureNamesStep(t *testing.T) {
	cfg := testConfig(t)
	boom := errors.New("boom")
	fake := &runner.Fake{OnRun: func(c runner.Call) error {
		if c.Args[0] == "venv" {
			return boom
		}
		return nil
	}}
	err := Run(cfg, fake)
	if !errors.Is(err, boom) || !strings.Contains(err.Error(), "uv venv") {
		t.Fatalf("err = %v, want wrapped boom mentioning uv venv", err)
	}
}

func TestRunUVInitFailureStops(t *testing.T) {
	cfg := testConfig(t)
	fake := &runner.Fake{OnRun: func(runner.Call) error { return errors.New("no uv") }}
	err := Run(cfg, fake)
	if err == nil || !strings.Contains(err.Error(), "uv init") {
		t.Fatalf("err = %v, want uv init error", err)
	}
	if len(fake.Calls) != 1 {
		t.Errorf("got %d calls, want 1", len(fake.Calls))
	}
}
