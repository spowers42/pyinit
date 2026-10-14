package preflight

import (
	"os"
	"path/filepath"
	"testing"
)

// A name unlikely to exist in /usr/local/bin or /opt/homebrew/bin on the test host.
const testTool = "pyinit-locate-test-tool"

func writeExecutable(t *testing.T, dir, name string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRealEnvLocateOnPath(t *testing.T) {
	bin := t.TempDir()
	want := writeExecutable(t, bin, testTool)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", bin)

	loc, ok := RealEnv{}.Locate(testTool)
	if !ok || loc != (Location{Path: want, OnPath: true}) {
		t.Fatalf("Locate = %+v, %v; want %q on PATH", loc, ok, want)
	}
}

func TestRealEnvLocateHomeFallback(t *testing.T) {
	for _, sub := range []string{".cargo/bin", ".local/bin"} {
		t.Run(sub, func(t *testing.T) {
			home := t.TempDir()
			want := writeExecutable(t, filepath.Join(home, sub), testTool)
			t.Setenv("HOME", home)
			t.Setenv("PATH", t.TempDir())

			loc, ok := RealEnv{}.Locate(testTool)
			if !ok || loc != (Location{Path: want, OnPath: false}) {
				t.Fatalf("Locate = %+v, %v; want %q off PATH", loc, ok, want)
			}
		})
	}
}

func TestRealEnvLocatePrefersPath(t *testing.T) {
	home := t.TempDir()
	writeExecutable(t, filepath.Join(home, ".local/bin"), testTool)
	bin := t.TempDir()
	want := writeExecutable(t, bin, testTool)
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin)

	loc, ok := RealEnv{}.Locate(testTool)
	if !ok || loc != (Location{Path: want, OnPath: true}) {
		t.Fatalf("Locate = %+v, %v; want %q on PATH", loc, ok, want)
	}
}

func TestRealEnvLocateNotFound(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir())

	if loc, ok := (RealEnv{}).Locate(testTool); ok {
		t.Fatalf("Locate = %+v, want not found", loc)
	}
}
