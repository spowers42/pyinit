package runner

import (
	"errors"
	"strings"
	"testing"
)

func TestResolvingRewritesMappedName(t *testing.T) {
	next := &Fake{}
	r := Resolving(next, map[string]string{"uv": "/home/u/.local/bin/uv"})
	if err := r.Run("/proj", "uv", "init", "--python", "3.12", "."); err != nil {
		t.Fatal(err)
	}
	if len(next.Calls) != 1 {
		t.Fatalf("calls = %v", next.Calls)
	}
	c := next.Calls[0]
	if c.Name != "/home/u/.local/bin/uv" {
		t.Errorf("name = %q", c.Name)
	}
	if c.Dir != "/proj" {
		t.Errorf("dir = %q", c.Dir)
	}
	if got := strings.Join(c.Args, " "); got != "init --python 3.12 ." {
		t.Errorf("args = %q", got)
	}
}

func TestResolvingPassesUnmappedNameThrough(t *testing.T) {
	next := &Fake{}
	r := Resolving(next, map[string]string{"uv": "/opt/uv"})
	if err := r.Run("", "git", "status"); err != nil {
		t.Fatal(err)
	}
	if got := next.Calls[0].String(); got != "git status" {
		t.Errorf("call = %q", got)
	}
}

func TestResolvingReturnsNextError(t *testing.T) {
	boom := errors.New("boom")
	next := &Fake{OnRun: func(Call) error { return boom }}
	r := Resolving(next, map[string]string{"uv": "/opt/uv"})
	if err := r.Run("", "uv"); !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
}
