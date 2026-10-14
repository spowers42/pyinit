package main

import "runtime/debug"

// resolveVersion returns the version to report. Release builds set it via
// ldflags; otherwise a `go install …@vX.Y.Z` build carries the module version
// in its build info. Builds without a module version ("(devel)") stay "dev".
func resolveVersion(ldflag string, readBuildInfo func() (*debug.BuildInfo, bool)) string {
	if ldflag != "dev" {
		return ldflag
	}
	info, ok := readBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return ldflag
	}
	return info.Main.Version
}
