package main

import (
	"runtime/debug"
	"testing"
)

func buildInfo(mainVersion string) func() (*debug.BuildInfo, bool) {
	return func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: mainVersion}}, true
	}
}

func noBuildInfo() (*debug.BuildInfo, bool) { return nil, false }

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name      string
		ldflag    string
		readBuild func() (*debug.BuildInfo, bool)
		want      string
	}{
		{"ldflag version wins", "v0.2.0", buildInfo("v0.1.0"), "v0.2.0"},
		{"go install of a tag reports the module version", "dev", buildInfo("v0.2.0"), "v0.2.0"},
		{"local build stays dev", "dev", buildInfo("(devel)"), "dev"},
		{"no build info stays dev", "dev", noBuildInfo, "dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveVersion(tt.ldflag, tt.readBuild); got != tt.want {
				t.Errorf("resolveVersion(%q) = %q, want %q", tt.ldflag, got, tt.want)
			}
		})
	}
}
