package scaffold

import (
	"slices"
	"testing"

	"github.com/spowers42/pyinit/internal/preflight"
)

// Every command scaffold runs must be a required tool, so preflight resolves it.
func TestRunOnlyRunsRequiredTools(t *testing.T) {
	fake := fakeUV()
	if err := Run(testConfig(t), fake); err != nil {
		t.Fatal(err)
	}
	if len(fake.Calls) == 0 {
		t.Fatal("expected Run to execute commands")
	}
	tools := preflight.ToolNames()
	for _, c := range fake.Calls {
		if !slices.Contains(tools, c.Name) {
			t.Errorf("scaffold runs %q, which is not a required tool %v", c.Name, tools)
		}
	}
}
