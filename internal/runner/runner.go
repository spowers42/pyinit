// Package runner is the seam for executing external commands.
package runner

import (
	"os"
	"os/exec"
	"strings"
)

// Runner runs a command in a directory.
type Runner interface {
	Run(dir, name string, args ...string) error
}

// Exec is the production Runner; it streams the command's output to the terminal.
type Exec struct{}

// Run executes name with args in dir, streaming stdout and stderr.
func (Exec) Run(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Resolving returns a Runner that runs each command named in paths from the
// mapped path instead, delegating to next. Names not in paths pass through unchanged.
func Resolving(next Runner, paths map[string]string) Runner {
	return resolving{next: next, paths: paths}
}

type resolving struct {
	next  Runner
	paths map[string]string
}

// Run executes name with args in dir via next, substituting name's mapped path if it has one.
func (r resolving) Run(dir, name string, args ...string) error {
	if p, ok := r.paths[name]; ok {
		name = p
	}
	return r.next.Run(dir, name, args...)
}

// Call records a single invocation made through a Fake.
type Call struct {
	Dir  string
	Name string
	Args []string
}

// String renders the call as a shell-like command line.
func (c Call) String() string {
	return strings.Join(append([]string{c.Name}, c.Args...), " ")
}

// Fake is a recording Runner for tests. It executes nothing.
type Fake struct {
	// Calls lists every invocation in order.
	Calls []Call
	// OnRun, if set, is invoked after recording each call; its error is returned.
	// Use it to mimic a command's side effects or to inject failures.
	OnRun func(Call) error
}

// Run records the call and delegates to OnRun if set.
func (f *Fake) Run(dir, name string, args ...string) error {
	c := Call{Dir: dir, Name: name, Args: args}
	f.Calls = append(f.Calls, c)
	if f.OnRun != nil {
		return f.OnRun(c)
	}
	return nil
}
