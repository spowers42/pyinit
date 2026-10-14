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

// isolate points HOME and PATH at fresh temp dirs and unsets XDG_BIN_HOME.
func isolate(t *testing.T) (home string) {
	t.Helper()
	home = t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", t.TempDir())
	t.Setenv("XDG_BIN_HOME", "")
	return home
}

func assertLocated(t *testing.T, want Location) {
	t.Helper()
	loc, ok := RealEnv{}.Locate(testTool)
	if !ok || loc != want {
		t.Fatalf("Locate = %+v, %v; want %+v", loc, ok, want)
	}
}

func TestRealEnvLocateOnPath(t *testing.T) {
	isolate(t)
	bin := t.TempDir()
	want := writeExecutable(t, bin, testTool)
	t.Setenv("PATH", bin)

	assertLocated(t, Location{Path: want, OnPath: true})
}

func TestRealEnvLocateHomeFallback(t *testing.T) {
	for _, sub := range []string{".cargo/bin", ".local/bin"} {
		t.Run(sub, func(t *testing.T) {
			home := isolate(t)
			want := writeExecutable(t, filepath.Join(home, sub), testTool)

			assertLocated(t, Location{Path: want, OnPath: false})
		})
	}
}

func TestRealEnvLocateXDGBinHome(t *testing.T) {
	isolate(t)
	xdg := t.TempDir()
	want := writeExecutable(t, xdg, testTool)
	t.Setenv("XDG_BIN_HOME", xdg)

	assertLocated(t, Location{Path: want, OnPath: false})
}

func TestRealEnvLocateSystemDirFallback(t *testing.T) {
	isolate(t)
	// Stand-ins for /usr/local/bin and /opt/homebrew/bin, which tests can't write to.
	usrLocal, homebrew := t.TempDir(), t.TempDir()
	orig := systemDirs
	systemDirs = []string{usrLocal, homebrew}
	t.Cleanup(func() { systemDirs = orig })

	for _, dir := range systemDirs {
		want := writeExecutable(t, dir, testTool)
		assertLocated(t, Location{Path: want, OnPath: false})
		if err := os.Remove(want); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRealEnvLocatePrefersPath(t *testing.T) {
	home := isolate(t)
	writeExecutable(t, filepath.Join(home, ".local/bin"), testTool)
	bin := t.TempDir()
	want := writeExecutable(t, bin, testTool)
	t.Setenv("PATH", bin)

	assertLocated(t, Location{Path: want, OnPath: true})
}

func TestRealEnvLocateNotFound(t *testing.T) {
	isolate(t)

	loc, ok := RealEnv{}.Locate(testTool)
	if ok {
		t.Fatalf("Locate = %+v, want not found", loc)
	}
}
